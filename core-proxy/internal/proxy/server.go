package proxy

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"sentinelx/core-proxy/internal/http1"
	"sentinelx/core-proxy/internal/http2"
	"sentinelx/core-proxy/internal/metrics"
	"sentinelx/core-proxy/internal/storage"
	tlsmanager "sentinelx/core-proxy/internal/tls"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Server is the core proxy server.
type Server struct {
	addr       string
	listener   net.Listener
	wg         sync.WaitGroup
	shutdown   chan struct{}
	tlsManager *tlsmanager.Manager
	store      *storage.Store
	tracer     trace.Tracer
}

// NewServer creates a new proxy server.
func NewServer(addr string) (*Server, error) {
	tlsManager, err := tlsmanager.NewManager()
	if err != nil {
		return nil, err
	}
	store, err := storage.NewStore("sentinelx.db")
	if err != nil {
		return nil, err
	}
	tracer := otel.Tracer("sentinelx/core-proxy")
	return &Server{
		addr:       addr,
		shutdown:   make(chan struct{}),
		tlsManager: tlsManager,
		store:      store,
		tracer:     tracer,
	}, nil
}

// Start starts the proxy server.
func (s *Server) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	log.Printf("Proxy server listening on %s", s.addr)

	s.wg.Add(1)
	go s.run()

	return nil
}

// run is the main loop for accepting connections.
func (s *Server) run() {
	defer s.wg.Done()
	for {
		select {
		case <-s.shutdown:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				// Check if the listener was closed.
				select {
				case <-s.shutdown:
					return
				default:
					log.Printf("Error accepting connection: %v", err)
				}
				continue
			}

			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

// handleConnection handles an incoming client connection.
func (s *Server) handleConnection(clientConn net.Conn) {
	metrics.ConnectionsTotal.Inc()
	defer s.wg.Done()
	defer clientConn.Close()

	log.Printf("Accepted connection from %s", clientConn.RemoteAddr())

	tlsConfig, err := s.TLSConfig()
	if err != nil {
		log.Printf("Failed to create TLS config: %v", err)
		return
	}

	tlsConn := tls.Server(clientConn, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		log.Printf("TLS handshake failed: %v", err)
		return
	}
	defer tlsConn.Close()

	// Determine the protocol to use.
	state := tlsConn.ConnectionState()
	switch state.NegotiatedProtocol {
	case "h2":
		processor := http2.NewProcessor(tlsConn)
		processor.Process()
	case "http/1.1":
		processor := http1.NewProcessor(tlsConn, s.store, s.tracer)
		processor.Process()
	default:
		// If no protocol is negotiated, we can default to HTTP/1.1 or close the connection.
		// For now, we'll default to HTTP/1.1.
		log.Printf("No protocol negotiated, defaulting to HTTP/1.1")
		processor := http1.NewProcessor(tlsConn, s.store, s.tracer)
		processor.Process()
	}
}

// Shutdown gracefully shuts down the proxy server.
func (s *Server) Shutdown(ctx context.Context) error {
	close(s.shutdown)
	if err := s.listener.Close(); err != nil {
		return err
	}
	if err := s.store.Close(); err != nil {
		return err
	}

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TLSConfig returns a TLS config that uses the TLS manager.
func (s *Server) TLSConfig() (*tls.Config, error) {
	return &tls.Config{
		GetCertificate: s.tlsManager.GetCertificate,
		NextProtos:     []string{"h2", "http/1.1"},
	}, nil
}
