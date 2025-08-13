package otel

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func Start(ctx context.Context, name string) (trace.Span, context.Context) {
	tr := otel.Tracer("core-proxy")
	ctx, span := tr.Start(ctx, name)
	return span, ctx
}

func Annotate(span trace.Span, key, val string) {
	span.SetAttributes(attribute.String(key, val))
}

func MarkErr(span trace.Span, err error) {
	span.RecordError(err)
	span.SetAttributes(attribute.String("error", err.Error()))
}
