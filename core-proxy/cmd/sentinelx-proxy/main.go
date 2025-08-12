package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sentinelx/core-proxy/config"
	"sentinelx/core-proxy/listener"
	"sentinelx/core-proxy/metrics"
	tlsmgr "sentinelx/core-proxy/tls"
)

var (
	configPath = flag.String("config", "core-proxy/config/example_config.yaml", "Path to config file")
)

func main() {
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("config load: %v", err)
	}

	if err := metrics.InitTracing(cfg.Tracing); err != nil {
		log.Fatalf("tracing init: %v", err)
	}
	defer metrics.ShutdownTracing(context.Background())

	tlsMgr, err := tlsmgr.NewManager(tlsmgr.Config{
		RootCertPath: cfg.TLS.RootCA.CertPath,
		RootKeyPath:  cfg.TLS.RootCA.KeyPath,
		DefaultTTL:   24 * time.Hour,
	})
	if err != nil {
		log.Fatalf("tls manager: %v", err)
	}

	// initialize metrics server (Prometheus)
	if err := metrics.Init(cfg.Metrics); err != nil {
		log.Fatalf("metrics init: %v", err)
	}
	go func() {
		if cfg.Metrics.Enabled {
			if err := metrics.Serve(cfg.Metrics); err != nil {
				log.Printf("metrics server error: %v", err)
			}
		}
	}()

	// create listener manager
	lm := listener.NewManager(cfg, tlsMgr)

	// start listeners
	if err := lm.StartAll(); err != nil {
		log.Fatalf("failed starting listeners: %v", err)
	}
	log.Printf("core-proxy started; listeners up")

	// graceful shutdown on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Printf("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Graceful.DrainTimeout.Duration)
	defer cancel()

	if err := lm.ShutdownAll(shutdownCtx); err != nil {
		log.Printf("error during shutdown: %v", err)
	}
	log.Printf("core-proxy stopped gracefully")
}
