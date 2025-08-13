package listener

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"sentinelx/core-proxy/config"
)

type ApplyPlan struct {
	Diff            DiffResult
	ApplyTimeout    time.Duration
	DrainTimeout    time.Duration
	RollbackOnFail  bool
	RotateLeafCache bool
}

type PlanExecutor interface {
	StartListener(ctx context.Context, l config.ListenerConfig) error
	UpdateListener(ctx context.Context, ch ChangedListener) error
	DrainAndStopListener(ctx context.Context, l config.ListenerConfig, deadline time.Duration) error
	HasListener(name string) bool
}

func (p ApplyPlan) Apply(ctx context.Context, exec PlanExecutor) error {
	ctx, cancel := context.WithTimeout(ctx, p.ApplyTimeout)
	defer cancel()

	var (
		started []config.ListenerConfig
		mu      sync.Mutex
		wg      sync.WaitGroup
		errOnce error
		setErr  = func(e error) { if errOnce == nil { errOnce = e } }
	)

	// Start added
	for _, add := range p.Diff.Added {
		lc := add
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := exec.StartListener(ctx, lc); err != nil {
				setErr(fmt.Errorf("start %s: %w", lc.Name, err))
			} else {
				mu.Lock()
				started = append(started, lc)
				mu.Unlock()
			}
		}()
	}

	// Update changed
	for _, ch := range p.Diff.Changed {
		choice := ch // capture
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := exec.UpdateListener(ctx, choice); err != nil {
				setErr(fmt.Errorf("update %s: %w", choice.New.Name, err))
			}
		}()
	}

	// Drain removed
	for _, rm := range p.Diff.Removed {
		rc := rm
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !exec.HasListener(rc.Name) {
				return
			}
			if err := exec.DrainAndStopListener(ctx, rc, p.DrainTimeout); err != nil {
				setErr(fmt.Errorf("drain %s: %w", rc.Name, err))
			}
		}()
	}

	wg.Wait()

	if errOnce != nil && p.RollbackOnFail {
		// Best-effort rollback: stop any newly started listeners.
		var rbWg sync.WaitGroup
		for _, s := range started {
			s := s
			rbWg.Add(1)
			go func() {
				defer rbWg.Done()
				_ = exec.DrainAndStopListener(context.Background(), s, 3*time.Second)
			}()
		}
		rbWg.Wait()
		return errOnce
	}
	return errOnce
}

var ErrInvalidPlan = errors.New("invalid plan")
