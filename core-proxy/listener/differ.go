package listener

import (
	"fmt"
	"reflect"

	"sentinelx/core-proxy/config"
)

// DiffResult: plan to apply atomically.
type DiffResult struct {
	Added   []config.ListenerConfig
	Changed []ChangedListener
	Removed []config.ListenerConfig
}

type ChangedListener struct {
	Old config.ListenerConfig
	New config.ListenerConfig
	// Hints for fast path updates:
	BindChanged         bool
	TLSChanged          bool
	BackpressureChanged bool
	RateLimitChanged    bool
	ALPNChanged         bool
	LoggingChanged      bool
}

func Diff(oldCfg, newCfg *config.Config) (DiffResult, error) {
	oldIdx := indexByName(oldCfg.Listeners)
	newIdx := indexByName(newCfg.Listeners)

	var res DiffResult

	// Added & Changed
	for name, newL := range newIdx {
		if oldL, ok := oldIdx[name]; !ok {
			res.Added = append(res.Added, newL)
		} else {
			ch := diffListener(oldL, newL)
			if ch != nil {
				res.Changed = append(res.Changed, *ch)
			}
		}
	}
	// Removed
	for name, oldL := range oldIdx {
		if _, ok := newIdx[name]; !ok {
			res.Removed = append(res.Removed, oldL)
		}
	}
	return res, nil
}

func indexByName(ls []config.ListenerConfig) map[string]config.ListenerConfig {
	m := make(map[string]config.ListenerConfig, len(ls))
	for _, l := range ls {
		m[l.Name] = l
	}
	return m
}

func diffListener(oldL, newL config.ListenerConfig) *ChangedListener {
	if reflect.DeepEqual(oldL, newL) {
		return nil
	}
	ch := &ChangedListener{Old: oldL, New: newL}
	ch.BindChanged = oldL.Bind != newL.Bind
	// This is a simplified diff. A real implementation would check sub-fields.
	ch.TLSChanged = !reflect.DeepEqual(oldL.TLS, newL.TLS)
	ch.BackpressureChanged = false // TODO: Implement this
	ch.RateLimitChanged = false    // TODO: Implement this
	ch.ALPNChanged = !reflect.DeepEqual(oldL.ALPN, newL.ALPN)
	ch.LoggingChanged = false      // TODO: Implement this
	return ch
}

func (d DiffResult) String() string {
	return fmt.Sprintf("added=%d changed=%d removed=%d", len(d.Added), len(d.Changed), len(d.Removed))
}
