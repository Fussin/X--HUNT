package health

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

type StatusSource interface {
	LivenessOK() bool
	Readiness() ReadinessReport
}

type HTTP struct {
	Source StatusSource
}

func (h *HTTP) Healthz(w http.ResponseWriter, r *http.Request) {
	ok := h.Source.LivenessOK()
	status := "ok"
	code := 200
	if !ok { status = "fail"; code = 500 }

	resp := map[string]any{
		"status":       status,
		"uptime_sec":   time.Since(startTime).Seconds(),
		"goroutines":   runtime.NumGoroutine(),
		"version":      buildVersion,
		"build":        buildCommit,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *HTTP) Readyz(w http.ResponseWriter, r *http.Request) {
	rep := h.Source.Readiness()
	code := 200
	if rep.Status != "ready" { code = 503 }
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(rep)
}

// set by main at init/build time
var (
	startTime    = time.Now()
	buildVersion = "dev"
	buildCommit  = "unknown"
)
