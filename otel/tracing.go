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

type TracerOptions struct {
	Method     TracingMethod
	Stdout     []stdouttrace.Option
	Grpc       []otlptracegrpc.Option
	Http       []otlptracehttp.Option
	BatchTimer time.Duration // the delay between batch posts. default is 5s
}

func SetupTracing(
	ctx context.Context,
	options TracerOptions,
	shutdownFuncs *ShutdownFuncs,
	res *resource.Resource,
) (*trace.TracerProvider, error) {
	// it might be useful to be able to send outputs to multiple streams.
	// sadly, not today.

	var traceExporter trace.SpanExporter
	var err error
	if options.Method == TracingMethodStdout {
		traceExporter, err = stdouttrace.New(options.Stdout...)
		if err != nil {
			return nil, err
		}
	} else if options.Method == TracingMethodGRPC {
		traceExporter, err = otlptracegrpc.New(ctx, options.Grpc...)
		if err != nil {
			return nil, err
		}
	} else if options.Method == TracingMethodHTTP {
		traceExporter, err = otlptracehttp.New(ctx, options.Http...)
		if err != nil {
			return nil, err
		}
	}

	var batchOptions []trace.BatchSpanProcessorOption
	if options.BatchTimer > 0 {
		batchOptions = append(
			batchOptions,
			trace.WithBatchTimeout(options.BatchTimer),
		)
	}
	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(
			traceExporter,
			batchOptions...,
		),
		trace.WithResource(res),
	)
	shutdownFuncs.append(tracerProvider.Shutdown)
	return tracerProvider, nil
}
