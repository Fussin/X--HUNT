package listener

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	cfg "sentinelx/core-proxy/config"
	"sentinelx/core-proxy/internal/metrics"
	tlsmgr "sentinelx/core-proxy/tls"
)

type Manager struct {
	cfg       *cfg.Config
	servers   map[string]*http.Server
	listeners map[string]net.Listener
	mtx       sync.Mutex
	wg        sync.WaitGroup

	tlsManager *tlsmgr.Manager
}

func NewManager(c *cfg.Config, tm *tlsmgr.Manager) *Manager {
	return &Manager{
		cfg:       c,
		servers:   make(map[string]*http.Server),
		listeners: make(map[string]net.Listener),
		tlsManager: tm,
	}
}

func (m *Manager) StartAll() error {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	if len(m.cfg.Listeners) == 0 {
		return errors.New("no listeners configured")
	}
	for i := range m.cfg.Listeners {
		lc := m.cfg.Listeners[i]
		mux := http.NewServeMux()
		// Replace with multiplexer later
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ctx := r.Context()
			// simple hello
			w.Header().Set("X-SentinelX", "core-proxy")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok\n"))

			// record metrics/tracing
			duration := time.Since(start).Seconds()
			metrics.RecordRequest(ctx, lc.Name, r.Method, http.StatusOK, duration)
		})

		srv := &http.Server{
			Handler:      mux,
			ReadTimeout:  time.Duration(lc.IdleTimeout.Duration),
			WriteTimeout: time.Duration(lc.IdleTimeout.Duration),
		}

		addr := lc.Bind
		var ln net.Listener
		var err error

		if lc.TLS.Enabled {
			// configure tls.Config with GetCertificate callback
			tcfg := &tls.Config{
				MinVersion: tls.VersionTLS12,
				NextProtos: lc.ALPN,
				GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
					// ask tlsManager to generate a cert for SNI (or use IP)
					name := hello.ServerName
					if name == "" {
						name = hello.Conn.RemoteAddr().String()
					}
					certPEM, keyPEM, err := m.tlsManager.GenerateLeafForHost([]string{name})
					if err != nil {
						return nil, err
					}
					cert, err := tls.X509KeyPair(certPEM, keyPEM)
					if err != nil { return nil, err }
					return &cert, nil
				},
			}
			ln, err = tls.Listen("tcp", addr, tcfg)
		} else {
			ln, err = net.Listen("tcp", addr)
		}
		if err != nil {
			return fmt.Errorf("bind %s: %w", addr, err)
		}

		m.servers[lc.Name] = srv
		m.listeners[lc.Name] = ln

		m.wg.Add(1)
		go func(name string, srv *http.Server, ln net.Listener, lc cfg.ListenerConfig) {
			defer m.wg.Done()
			log.Printf("listener %s serving on %s (tls=%v)", name, ln.Addr(), lc.TLS.Enabled)
			if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
				log.Printf("listener %s error: %v", name, err)
			}
		}(lc.Name, srv, ln, lc)
	}
	return nil
}

func (m *Manager) ShutdownAll(ctx context.Context) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	var errs []error
	for name, srv := range m.servers {
		shCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		if err := srv.Shutdown(shCtx); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
		}
		cancel()
	}
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}
	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}
