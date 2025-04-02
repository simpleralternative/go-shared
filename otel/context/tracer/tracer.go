package tracer

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type tracerKeyType int

const tracerKey tracerKeyType = 0

func Set(ctx context.Context, trc trace.TracerProvider) context.Context {
	return context.WithValue(ctx, tracerKey, trc)
}

func Get(ctx context.Context) trace.TracerProvider {
	value := ctx.Value(tracerKey)
	if value == nil {
		return otel.GetTracerProvider()
	}
	trc, ok := value.(trace.TracerProvider)
	if !ok {
		return otel.GetTracerProvider()
	}
	return trc
}
