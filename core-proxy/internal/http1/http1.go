package http1

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"sentinelx/core-proxy/internal/metrics"
	"sentinelx/core-proxy/internal/storage"
	"sentinelx/core-proxy/internal/transform"
	"sentinelx/core-proxy/internal/websocket"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
)

// Processor handles a single HTTP/1.1 connection.
type Processor struct {
	clientConn  net.Conn
	transformer *transform.Transformer
	store       *storage.Store
	tracer      trace.Tracer
}

// NewProcessor creates a new HTTP/1.1 processor.
func NewProcessor(clientConn net.Conn, store *storage.Store, tracer trace.Tracer) *Processor {
	// For now, we'll create a simple transformer here.
	// In the future, this should be configurable.
	transformer := transform.NewTransformer(
		&transform.JSONInterceptor{
			ModifyFunc: func(data map[string]interface{}) (map[string]interface{}, error) {
				data["modified"] = true
				return data, nil
			},
		},
	)
	return &Processor{
		clientConn:  clientConn,
		transformer: transformer,
		store:       store,
		tracer:      tracer,
	}
}

// Process handles the HTTP/1.1 connection.
func (p *Processor) Process() {
	br := bufio.NewReader(p.clientConn)

	for {
		req, err := http.ReadRequest(br)
		if err != nil {
			// TODO: Handle errors properly (e.g., EOF, timeout).
			log.Printf("Failed to read request: %v", err)
			return
		}

		// Check for WebSocket upgrade.
		if isWebSocketUpgrade(req) {
			p.handleWebSocket(req)
			return // WebSocket processor takes over the connection.
		}

		_, span := p.tracer.Start(context.Background(), "http1.request")
		defer span.End()

		metrics.RequestsTotal.Inc()
		timer := prometheus.NewTimer(metrics.RequestDuration)
		defer timer.ObserveDuration()

		// Intercept the request.
		if err := p.transformer.InterceptRequest(req); err != nil {
			log.Printf("Failed to intercept request: %v", err)
			return
		}

		// Connect to the target server.
		targetConn, err := net.Dial("tcp", req.Host)
		if err != nil {
			log.Printf("Failed to connect to target server %s: %v", req.Host, err)
			resp := &http.Response{
				StatusCode: http.StatusBadGateway,
				Body:       io.NopCloser(bytes.NewBufferString("Bad Gateway")),
			}
			resp.Write(p.clientConn)
			return
		}
		defer targetConn.Close()

		// Send the request to the target server.
		if err := req.Write(targetConn); err != nil {
			log.Printf("Failed to write request to target: %v", err)
			return
		}

		// Read the response from the target server.
		targetBr := bufio.NewReader(targetConn)
		resp, err := http.ReadResponse(targetBr, req)
		if err != nil {
			log.Printf("Failed to read response from target: %v", err)
			return
		}
		defer resp.Body.Close()

		// Intercept the response.
		if err := p.transformer.InterceptResponse(resp); err != nil {
			log.Printf("Failed to intercept response: %v", err)
			return
		}

		// Save the request and response.
		reqBytes, err := httputil.DumpRequest(req, true)
		if err != nil {
			log.Printf("Failed to dump request: %v", err)
		}
		respBytes, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Printf("Failed to dump response: %v", err)
		}
		if reqBytes != nil && respBytes != nil {
			entry := storage.Entry{
				Request:  storage.Request{Raw: reqBytes},
				Response: storage.Response{Raw: respBytes},
			}
			// TODO: Use a real session ID.
			session := &storage.Session{
				ID:      []byte("test-session"),
				Entries: []storage.Entry{entry},
			}
			if err := p.store.SaveSession(session); err != nil {
				log.Printf("Failed to save session: %v", err)
			}
		}

		// Send the response back to the client.
		if err := resp.Write(p.clientConn); err != nil {
			log.Printf("Failed to write response to client: %v", err)
			return
		}

		// Handle client keep-alive.
		if req.Close {
			return
		}
	}
}

func isWebSocketUpgrade(req *http.Request) bool {
	return strings.ToLower(req.Header.Get("Upgrade")) == "websocket" &&
		strings.Contains(strings.ToLower(req.Header.Get("Connection")), "upgrade")
}

func (p *Processor) handleWebSocket(req *http.Request) {
	log.Printf("Handling WebSocket upgrade for %s", req.Host)

	targetConn, err := net.Dial("tcp", req.Host)
	if err != nil {
		log.Printf("Failed to connect to target server %s: %v", req.Host, err)
		return
	}

	// Send the upgrade request.
	if err := req.Write(targetConn); err != nil {
		log.Printf("Failed to write upgrade request to target: %v", err)
		targetConn.Close()
		return
	}

	// Read the response.
	br := bufio.NewReader(targetConn)
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		log.Printf("Failed to read upgrade response from target: %v", err)
		targetConn.Close()
		return
	}

	// Check if the upgrade was successful.
	if resp.StatusCode != http.StatusSwitchingProtocols {
		log.Printf("WebSocket upgrade failed with status: %s", resp.Status)
		// Forward the response to the client and close.
		if err := resp.Write(p.clientConn); err != nil {
			log.Printf("Failed to write non-upgrade response to client: %v", err)
		}
		targetConn.Close()
		return
	}

	// Forward the 101 response to the client.
	if err := resp.Write(p.clientConn); err != nil {
		log.Printf("Failed to write upgrade response to client: %v", err)
		targetConn.Close()
		return
	}

	// Hand over the connections to the WebSocket processor.
	wsProcessor := websocket.NewProcessor(p.clientConn, targetConn)
	wsProcessor.Process()
}
