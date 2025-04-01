package otel

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

type LoggingMethod uint8

const (
	LoggingMethodStdout LoggingMethod = iota
	LoggingMethodGRPC
	LoggingMethodHTTP
)

type LoggerOptions struct {
	Method LoggingMethod
	Stdout []stdoutlog.Option
	Grpc   []otlploggrpc.Option
	Http   []otlploghttp.Option
}

func SetupLogging(
	ctx context.Context,
	options LoggerOptions,
	shutdownFunc *ShutdownFuncs,
	res *resource.Resource,
) (*log.LoggerProvider, error) {
	var logExporter log.Exporter
	var err error
	if options.Method == LoggingMethodStdout {
		logExporter, err = stdoutlog.New(options.Stdout...)
		if err != nil {
			return nil, err
		}
	} else if options.Method == LoggingMethodGRPC {
		logExporter, err = otlploggrpc.New(ctx, options.Grpc...)
		if err != nil {
			return nil, err
		}
	} else if options.Method == LoggingMethodHTTP {
		logExporter, err = otlploghttp.New(ctx, options.Http...)
		if err != nil {
			return nil, err
		}
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
		log.WithResource(res),
	)
	shutdownFunc.append(loggerProvider.Shutdown)
	return loggerProvider, nil
}
