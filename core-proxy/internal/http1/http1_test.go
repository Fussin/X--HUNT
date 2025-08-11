package http1

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sentinelx/core-proxy/internal/storage"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel/trace"
)

func newTestStore(t *testing.T) *storage.Store {
	f, err := os.CreateTemp("", "sentinelx-test.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	t.Cleanup(func() {
		os.Remove(f.Name())
	})
	store, err := storage.NewStore(f.Name())
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	t.Cleanup(func() {
		store.Close()
	})
	return store
}

func newTestProcessor(clientConn net.Conn, store *storage.Store) *Processor {
	return NewProcessor(clientConn, store, trace.NewNoopTracerProvider().Tracer("test"))
}

func TestProcessor(t *testing.T) {
	// Create a mock target server.
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer targetServer.Close()

	// Create a pipe to simulate the client connection.
	clientConn, serverConn := net.Pipe()

	// Create the processor with the server side of the pipe.
	processor := newTestProcessor(serverConn, newTestStore(t))
	go func() {
		defer serverConn.Close()
		processor.Process()
	}()

	// --- Client side ---
	// Create a request that will be sent to the processor.
	// The host of the request should be the address of the target server.
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Host = targetServer.Listener.Addr().String()

	// Write the request to the client side of thepipe.
	if err := req.Write(clientConn); err != nil {
		t.Fatalf("Failed to write request: %v", err)
	}

	// Read the response from the client side of the pipe.
	br := bufio.NewReader(clientConn)
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	defer resp.Body.Close()

	// Verify the response.
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %s", resp.Status)
	}
}

func TestErrorHandling(t *testing.T) {
	// Create a pipe to simulate the client connection.
	clientConn, serverConn := net.Pipe()

	// Create the processor with the server side of the pipe.
	processor := newTestProcessor(serverConn, newTestStore(t))
	go func() {
		defer serverConn.Close()
		processor.Process()
	}()

	// --- Client side ---
	// Create a request that will fail to connect.
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Host = "127.0.0.1:1" // A port that is unlikely to be open.

	// Write the request to the client side of the pipe.
	if err := req.Write(clientConn); err != nil {
		t.Fatalf("Failed to write request: %v", err)
	}

	// Read the response from the client side of the pipe.
	br := bufio.NewReader(clientConn)
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("Expected status Bad Gateway, got %s", resp.Status)
	}
}

var upgrader = websocket.Upgrader{}

func TestWebSocketUpgrade(t *testing.T) {
	// Create a mock target WebSocket server.
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("Upgrade failed: %v", err)
			return
		}
		defer conn.Close()
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if err := conn.WriteMessage(mt, message); err != nil {
				break
			}
		}
	}))
	defer targetServer.Close()

	// Create a pipe to simulate the client connection.
	clientConn, serverConn := net.Pipe()

	// Create the processor with the server side of the pipe.
	processor := newTestProcessor(serverConn, newTestStore(t))
	go func() {
		defer serverConn.Close()
		processor.Process()
	}()

	// --- Client side ---
	// Create a WebSocket client that connects through the pipe.
	// We need to construct the URL so that the host is the target server's address.
	u, err := url.Parse(targetServer.URL)
	if err != nil {
		t.Fatalf("Failed to parse URL: %v", err)
	}
	u.Scheme = "ws"

	// The dialer needs a custom net.Conn.
	dialer := websocket.Dialer{
		NetDial: func(network, addr string) (net.Conn, error) {
			return clientConn, nil
		},
	}

	// Connect to the server.
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	// Send a message and check for the echo.
	if err := conn.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	_, p, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if string(p) != "hello" {
		t.Errorf("Expected 'hello', got '%s'", string(p))
	}
}

func TestChunkedRequest(t *testing.T) {
	// Create a mock target server that checks for chunked encoding.
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TransferEncoding == nil || len(r.TransferEncoding) == 0 || r.TransferEncoding[0] != "chunked" {
			t.Error("Expected chunked transfer encoding")
		}
		fmt.Fprintln(w, "Hello, client")
	}))
	defer targetServer.Close()

	// Create a pipe to simulate the client connection.
	clientConn, serverConn := net.Pipe()

	// Create the processor with the server side of the pipe.
	processor := newTestProcessor(serverConn, newTestStore(t))
	go func() {
		defer serverConn.Close()
		processor.Process()
	}()

	// --- Client side ---
	// Create a request with a chunked body.
	body := "Hello, world!"
	req, err := http.NewRequest("POST", "/", strings.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Host = targetServer.Listener.Addr().String()
	req.TransferEncoding = []string{"chunked"}

	// Write the request to the client side of the pipe.
	if err := req.Write(clientConn); err != nil {
		t.Fatalf("Failed to write request: %v", err)
	}

	// Read the response from the client side of the pipe.
	br := bufio.NewReader(clientConn)
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %s", resp.Status)
	}
}
