package main

import (
	"context"
	"time"

	"github.com/simpleralternative/go-shared/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func main() {
	ctx := context.Background()
	sdf := &otel.ShutdownFuncs{}
	tracerProvider, err := otel.SetupTracing(
		ctx,
		otel.TracerOptions{
			Method:     otel.TracingMethodStdout,
			Stdout:     []stdouttrace.Option{stdouttrace.WithPrettyPrint()},
			BatchTimer: 10 * time.Millisecond, // unrealistically short
		},
		sdf,
		// this overrides OTEL_SERVICE_NAME
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("tracer example"),
		),
	)
	if err != nil {
		defer sdf.Shutdown(ctx)
	}
	// solve the async send problem
	defer tracerProvider.ForceFlush(ctx)

	tracer := tracerProvider.Tracer("main")
	_, span := tracer.Start(ctx, "example")
	defer span.End() // the normal pattern is to End at close at end of scope

	// some content to send
	span.AddEvent("start")
	time.Sleep(1 * time.Second)
	span.AddEvent("end")

	// manually triggering End safely closes the span now and the defer is noop
	span.End()
}
