package proxy

import (
	"context"
	"crypto/tls"
	"sentinelx/core-proxy/config"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	cfg := &config.Config{
		Listeners: []config.ListenerConfig{
			{Name: "test", Bind: "127.0.0.1:0"},
		},
	}
	server, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give the server a moment to start.
	time.Sleep(100 * time.Millisecond)

	// Attempt to connect to the server with a TLS client.
	// We need to skip verification because we are using a self-signed CA.
	conf := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "localhost",
	}

	conn, err := tls.Dial("tcp", server.listeners[0].Addr().String(), conf)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	conn.Close()

	// Shutdown the server.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("Server shutdown failed: %v", err)
	}
}
