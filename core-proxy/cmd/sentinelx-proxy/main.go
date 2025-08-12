package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sentinelx/core-proxy/config"
	"sentinelx/core-proxy/internal/metrics"
	"sentinelx/core-proxy/internal/proxy"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	cfgPath := os.Getenv("COREPROXY_CONFIG")
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	tp, err := newTracerProvider()
	if err != nil {
		log.Fatalf("Failed to create tracer provider: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	server, err := proxy.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Start the metrics server.
	if cfg.Metrics.Enabled {
		go metrics.Serve(cfg.Metrics.Bind)
	}

	// Wait for shutdown signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Use graceful shutdown from config.
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Graceful.DrainTimeout.Duration)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server gracefully stopped")
}

func newTracerProvider() (*sdktrace.TracerProvider, error) {
	// TODO: Configure tracer from config.
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}
