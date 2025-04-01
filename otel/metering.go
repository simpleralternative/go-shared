package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

type MetricMethod uint8

const (
	MetricMethodStdout MetricMethod = iota
	MetricMethodGRPC
	MetricMethodHTTP
)

type MetricOptions struct {
	Method     MetricMethod
	Stdout     []stdoutmetric.Option
	Grpc       []otlpmetricgrpc.Option
	Http       []otlpmetrichttp.Option
	BatchTimer time.Duration
}

func SetupMetrics(
	ctx context.Context,
	options MetricOptions,
	shutdownFuncs *ShutdownFuncs,
	res *resource.Resource,
) (*metric.MeterProvider, error) {
	var metricExporter metric.Exporter
	var err error
	if options.Method == MetricMethodStdout {
		metricExporter, err = stdoutmetric.New(options.Stdout...)
		if err != nil {
			return nil, err
		}
	} else if options.Method == MetricMethodGRPC {
		metricExporter, err = otlpmetricgrpc.New(ctx, options.Grpc...)
		if err != nil {
			return nil, err
		}
	} else if options.Method == MetricMethodHTTP {
		metricExporter, err = otlpmetrichttp.New(ctx, options.Http...)
		if err != nil {
			return nil, err
		}
	}

	meterProvider, err := metric.NewMeterProvider(
		metric.WithReader(
			metric.NewPeriodicReader(
				metricExporter,
				metric.WithInterval(options.BatchTimer),
			),
		),
		metric.WithResource(res),
	), nil

	shutdownFuncs.append(meterProvider.Shutdown)
	return meterProvider, nil
}
