package otel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

type TracingMethod uint8

const (
	TracingMethodStdout TracingMethod = iota
	TracingMethodGRPC
	TracingMethodHTTP
)

type TracerConfiguration struct {
	Method     TracingMethod
	Stdout     []stdouttrace.Option
	Grpc       []otlptracegrpc.Option
	Http       []otlptracehttp.Option
	BatchTimer time.Duration // the delay between batch posts. default is 5s
}

func SetupTracing(
	ctx context.Context,
	configuration *TracerConfiguration,
	shutdownFuncs *ShutdownFuncs,
	res *resource.Resource,
) (*trace.TracerProvider, error) {
	if configuration == nil {
		return nil, ErrConfigurationRequired
	}

	var traceExporter trace.SpanExporter
	var err error
	if configuration.Method == TracingMethodStdout {
		traceExporter, err = stdouttrace.New(configuration.Stdout...)
		if err != nil {
			return nil, err
		}
	} else if configuration.Method == TracingMethodGRPC {
		traceExporter, err = otlptracegrpc.New(ctx, configuration.Grpc...)
		if err != nil {
			return nil, err
		}
	} else if configuration.Method == TracingMethodHTTP {
		traceExporter, err = otlptracehttp.New(ctx, configuration.Http...)
		if err != nil {
			return nil, err
		}
	}

	var batchOptions []trace.BatchSpanProcessorOption
	if configuration.BatchTimer > 0 {
		batchOptions = append(
			batchOptions,
			trace.WithBatchTimeout(configuration.BatchTimer),
		)
	}
	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(
			traceExporter,
			batchOptions...,
		),
		trace.WithResource(res),
	)

	shutdownFuncs.append(tracerProvider.ForceFlush)
	shutdownFuncs.append(tracerProvider.Shutdown)
	return tracerProvider, nil
}
