package listener_test

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	cfg "sentinelx/core-proxy/config"
	"sentinelx/core-proxy/listener"
	tlsmgr "sentinelx/core-proxy/tls"
	"github.com/stretchr/testify/require"
)

func makeTestConfig(bind string, tlsEnabled bool) *cfg.Config {
	return &cfg.Config{
		Listeners: []cfg.ListenerConfig{
			{
				Name:       "test",
				Bind:       bind,
				Protocol:   "http",
				Backlog:    128,
				TCPNoDelay: true,
				TLS: cfg.ListenerTLS{
					Enabled: tlsEnabled,
				},
				IdleTimeout: cfg.Duration{Duration: 5 * time.Second},
			},
		},
		TLS: cfg.TLSConfig{ // empty root, manager will generate root
		},
		Graceful: cfg.GracefulCfg{
			DrainTimeout: cfg.Duration{Duration: 5 * time.Second},
		},
		Metrics: cfg.MetricsConfig{Enabled: false},
	}
}

// findFreePort returns a free TCP port for testing using :0 trick
func findFreePort(t *testing.T) string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	l.Close()
	return addr
}

func TestManager_StartShutdown_NoTLS(t *testing.T) {
	addr := findFreePort(t)
	cfg := makeTestConfig(addr, false)

	tm, err := tlsmgr.NewManager(tlsmgr.Config{DefaultTTL: time.Hour})
	require.NoError(t, err)

	m := listener.NewManager(cfg, tm)
	err = m.StartAll()
	require.NoError(t, err)

	// try an HTTP GET
	httpClient := &http.Client{Timeout: 2 * time.Second}
	resp, err := httpClient.Get("http://" + addr + "/")
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.Equal(t, "ok\n", string(body))

	// shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = m.ShutdownAll(ctx)
	require.NoError(t, err)
}

func TestManager_StartShutdown_WithTLS(t *testing.T) {
	addr := findFreePort(t)
	cfg := makeTestConfig(addr, true)

	// create manager - it will generate root CA
	tm, err := tlsmgr.NewManager(tlsmgr.Config{DefaultTTL: time.Hour})
	require.NoError(t, err)

	m := listener.NewManager(cfg, tm)
	err = m.StartAll()
	require.NoError(t, err)

	// Because server uses self-signed certs created by manager, use HTTP client that skips verify
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // nolint: gosec
	}
	httpClient := &http.Client{Transport: tr, Timeout: 3 * time.Second}
	resp, err := httpClient.Get("https://" + addr + "/")
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.Equal(t, "ok\n", string(body))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = m.ShutdownAll(ctx)
	require.NoError(t, err)
}
