package main

import (
	"context"

	"github.com/simpleralternative/go-shared/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func main() {
	ctx := context.Background()
	sdf := &otel.ShutdownFuncs{}
	metricProvider, err := otel.SetupMetrics(
		ctx,
		&otel.MeterConfiguration{
			Method: otel.MetricMethodStdout,
			Stdout: []stdoutmetric.Option{stdoutmetric.WithPrettyPrint()},
			// BatchTimer: 5 * time.Second, // default 5s
		},
		sdf,
		// this overrides OTEL_SERVICE_NAME
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("meter example"),
		),
	)
	if err != nil {
		panic(err)
	}
	defer sdf.Shutdown(ctx)

	metric := metricProvider.Meter("main")
	counter, err := metric.Int64Counter("counter")
	if err != nil {
		panic(err)
	}
	gauge, err := metric.Int64Gauge("gauge")
	if err != nil {
		panic(err)
	}
	histogram, err := metric.Int64Histogram("histogram")
	if err != nil {
		panic(err)
	}

	for id := range 100 {
		id64 := int64(id)
		counter.Add(ctx, id64)
		gauge.Record(ctx, id64)
		histogram.Record(ctx, id64)
	}
}
