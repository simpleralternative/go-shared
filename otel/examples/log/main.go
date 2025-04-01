package main

import (
	"context"
	"log/slog"

	"github.com/simpleralternative/go-shared/otel"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func main() {
	ctx := context.Background()
	sdf := &otel.ShutdownFuncs{}
	logProvider, err := otel.SetupLogging(
		ctx,
		otel.LoggerOptions{
			Method: otel.LoggingMethodStdout,
			Stdout: []stdoutlog.Option{stdoutlog.WithPrettyPrint()},
		},
		sdf,
		// this overrides OTEL_SERVICE_NAME
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("logger example"),
		),
	)
	if err != nil {
		defer sdf.Shutdown(ctx)
	}
	// the batch processor is async, on a timer, so we need to wait for output
	// or force it
	defer logProvider.ForceFlush(ctx)

	logger := slog.New(
		otelslog.NewHandler(
			"main",
			otelslog.WithLoggerProvider(logProvider),
		),
	)
	logger.ErrorContext(ctx, "some content to send")
}
