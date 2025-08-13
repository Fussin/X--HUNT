package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	backpressureEvents = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "coreproxy_backpressure_events_total",
			Help: "Backpressure decisions taken by policy engine",
		},
		[]string{"policy"},
	)
)

func init() {
	prometheus.MustRegister(backpressureEvents)
}

func IncBackpressureEvent(policy string) {
	backpressureEvents.WithLabelValues(policy).Inc()
}
