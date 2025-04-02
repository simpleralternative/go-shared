package otel

import (
	"context"

	"github.com/simpleralternative/go-shared/otel/context/logger"
	"github.com/simpleralternative/go-shared/otel/context/meter"
	"github.com/simpleralternative/go-shared/otel/context/tracer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(
	ctx context.Context,
	loggerOptions *LoggerConfiguration,
	meterOptions *MeterConfiguration,
	tracerOptions *TracerConfiguration,
	serviceName string,
) (context.Context, func(context.Context) error, error) {
	shutdownFuncs := &ShutdownFuncs{}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// may be overridden by environment: OTEL_SERVICE_NAME
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
	)

	if loggerOptions != nil {
		loggerProvider, err := SetupLogging(ctx, loggerOptions, shutdownFuncs, res)
		if err != nil {
			return shutdownFuncs.handleErr(ctx, err)
		}
		ctx = logger.Set(ctx, loggerProvider)
	}

	if meterOptions != nil {
		meterProvider, err := SetupMetrics(ctx, meterOptions, shutdownFuncs, res)
		if err != nil {
			return shutdownFuncs.handleErr(ctx, err)
		}
		ctx = meter.Set(ctx, meterProvider)
	}

	if tracerOptions != nil {
		tracerProvider, err := SetupTracing(ctx, tracerOptions, shutdownFuncs, res)
		if err != nil {
			return shutdownFuncs.handleErr(ctx, err)
		}
		ctx = tracer.Set(ctx, tracerProvider)
	}

	return ctx, shutdownFuncs.Shutdown, nil
}
