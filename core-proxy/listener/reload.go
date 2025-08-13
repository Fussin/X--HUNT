package listener

import (
	"context"
	"sync"
	"time"

	"sentinelx/core-proxy/config"
	"sentinelx/core-proxy/otel"
)

type ManagerReloader interface {
	GetConfig() *config.Config               // returns atomic snapshot
	SetConfig(*config.Config)                // atomic swap
	Executor() PlanExecutor                  // concrete executor
}

type Reloader struct {
	mu       sync.Mutex // serialize reloads
	manager  ManagerReloader
}

func NewReloader(m ManagerReloader) *Reloader { return &Reloader{manager: m} }

func (r *Reloader) Reload(ctx context.Context, newCfg *config.Config) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	span, ctx := otel.Start(ctx, "listener.reload")
	defer span.End()

	oldCfg := r.manager.GetConfig()

	diff, err := Diff(oldCfg, newCfg)
	if err != nil {
		otel.MarkErr(span, err)
		return err
	}

	plan := ApplyPlan{
		Diff:           diff,
		ApplyTimeout:   10 * time.Second, // TODO: Make this configurable.
		DrainTimeout:   oldCfg.Graceful.DrainTimeout.Duration,
		RollbackOnFail: oldCfg.HotReload.Enabled, // A reasonable default.
	}

	if err := plan.Apply(ctx, r.manager.Executor()); err != nil {
		otel.MarkErr(span, err)
		return err
	}

	// Swap config atomically only after successful apply.
	r.manager.SetConfig(newCfg)
	otel.Annotate(span, "diff", diff.String())
	return nil
}
