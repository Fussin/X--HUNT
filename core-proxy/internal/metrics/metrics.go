package metrics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	cfg "sentinelx/core-proxy/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

var (
	reqCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "coreproxy_requests_total",
		Help: "Total requests processed",
	}, []string{"listener", "method", "status"})
	reqDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "coreproxy_request_duration_seconds",
		Help:    "Latency in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"listener"})
	onceSync = false

	tracer oteltrace.Tracer
	tp     *sdktrace.TracerProvider
)

func Init(c cfg.MetricsConfig) error {
	if !c.Enabled {
		return nil
	}
	if !onceSync {
		prometheus.MustRegister(reqCounter, reqDuration)
		onceSync = true
	}
	return nil
}

func Serve(c cfg.MetricsConfig) error {
	http.Handle(c.Path, promhttp.Handler())
	return http.ListenAndServe(c.Bind, nil)
}

// InitTracing sets up OTLP HTTP exporter if tracing config is present.
// call this early (main) with cfg.Tracing
func InitTracing(tc cfg.TracingConfig) error {
	if !tc.Enabled || tc.OTLPEndpoint == "" {
		// tracing disabled; use noop tracer
		tracer = otel.Tracer("core-proxy")
		return nil
	}
	exp, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(tc.OTLPEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("failed create otlp exporter: %w", err)
	}
	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("sentinelx-core-proxy"),
		),
	)
	tp = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	tracer = otel.Tracer("core-proxy")
	return nil
}

// ShutdownTracing flushes exporter
func ShutdownTracing(ctx context.Context) error {
	if tp == nil {
		return nil
	}
	return tp.Shutdown(ctx)
}

// RecordRequest records Prometheus metrics and starts an OpenTelemetry span
func RecordRequest(ctx context.Context, listener, method string, status int, duration float64) {
	// metrics
	reqCounter.WithLabelValues(listener, method, fmt.Sprintf("%d", status)).Inc()
	reqDuration.WithLabelValues(listener).Observe(duration)

	// tracing
	if tracer == nil {
		tracer = otel.Tracer("core-proxy")
	}
	var span oteltrace.Span
	ctx, span = tracer.Start(ctx, "http.request",
		oteltrace.WithAttributes(
			attribute.String("listener", listener),
			attribute.String("http.method", method),
			attribute.Int("http.status_code", status),
			attribute.Float64("http.duration_s", duration),
		),
	)
	// end immediately (we're just recording metadata here); other parts may create child spans.
	span.End()
}
