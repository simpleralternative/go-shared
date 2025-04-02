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

type MeterConfiguration struct {
	Method     MetricMethod
	Stdout     []stdoutmetric.Option
	Grpc       []otlpmetricgrpc.Option
	Http       []otlpmetrichttp.Option
	BatchTimer time.Duration
}

func SetupMetrics(
	ctx context.Context,
	configuration *MeterConfiguration,
	shutdownFuncs *ShutdownFuncs,
	res *resource.Resource,
) (*metric.MeterProvider, error) {
	if configuration == nil {
		return nil, ErrConfigurationRequired
	}

	var metricExporter metric.Exporter
	var err error
	if configuration.Method == MetricMethodStdout {
		metricExporter, err = stdoutmetric.New(configuration.Stdout...)
		if err != nil {
			return nil, err
		}
	} else if configuration.Method == MetricMethodGRPC {
		metricExporter, err = otlpmetricgrpc.New(ctx, configuration.Grpc...)
		if err != nil {
			return nil, err
		}
	} else if configuration.Method == MetricMethodHTTP {
		metricExporter, err = otlpmetrichttp.New(ctx, configuration.Http...)
		if err != nil {
			return nil, err
		}
	}

	meterProvider, err := metric.NewMeterProvider(
		metric.WithReader(
			metric.NewPeriodicReader(
				metricExporter,
				metric.WithInterval(configuration.BatchTimer),
			),
		),
		metric.WithResource(res),
	), nil

	shutdownFuncs.append(meterProvider.ForceFlush)
	shutdownFuncs.append(meterProvider.Shutdown)
	return meterProvider, nil
}
