package admin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"sentinelx/core-proxy/config"
	"sentinelx/core-proxy/listener"
)

type Reloader interface {
	Reload(rctx context.Context, newCfg *config.Config) error
}

type Server struct {
	Auth   Middleware
	Load   func([]byte) (*config.Config, error)
	Reload func(rctx context.Context, cfg *config.Config) error
}

func (s *Server) PostReload(w http.ResponseWriter, r *http.Request) {
	if !s.Auth.Allow(w, r) {
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	cfg, err := s.Load(body)
	if err != nil {
		http.Error(w, "invalid config: "+err.Error(), 400)
		return
	}

	if err := s.Reload(r.Context(), cfg); err != nil {
		http.Error(w, "apply failed: "+err.Error(), 409)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
	})
}

// Optional dry-run: returns diff only
func (s *Server) PostValidate(w http.ResponseWriter, r *http.Request, current *config.Config) {
	if !s.Auth.Allow(w, r) {
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	newCfg, err := s.Load(body)
	if err != nil {
		http.Error(w, "invalid config: "+err.Error(), 400)
		return
	}

	diff, err := listener.Diff(current, newCfg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(diff)
}
