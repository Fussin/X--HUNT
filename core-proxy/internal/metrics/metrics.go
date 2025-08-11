package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

var (
	ConnectionsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sentinelx_connections_total",
			Help: "Total number of connections.",
		},
	)
	RequestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sentinelx_requests_total",
			Help: "Total number of requests.",
		},
	)
	RequestDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "sentinelx_request_duration_seconds",
			Help:    "Duration of requests.",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func init() {
	prometheus.MustRegister(ConnectionsTotal)
	prometheus.MustRegister(RequestsTotal)
	prometheus.MustRegister(RequestDuration)
}

// Serve serves the metrics endpoint on the given address.
func Serve(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Metrics server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start metrics server: %v", err)
	}
}
