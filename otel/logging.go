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

type LoggerConfiguration struct {
	Method LoggingMethod
	Stdout []stdoutlog.Option
	Grpc   []otlploggrpc.Option
	Http   []otlploghttp.Option
}

func SetupLogging(
	ctx context.Context,
	configuration *LoggerConfiguration,
	shutdownFunc *ShutdownFuncs,
	res *resource.Resource,
) (*log.LoggerProvider, error) {
	if configuration == nil {
		return nil, ErrConfigurationRequired
	}

	var logExporter log.Exporter
	var err error
	if configuration.Method == LoggingMethodStdout {
		logExporter, err = stdoutlog.New(configuration.Stdout...)
		if err != nil {
			return nil, err
		}
	} else if configuration.Method == LoggingMethodGRPC {
		logExporter, err = otlploggrpc.New(ctx, configuration.Grpc...)
		if err != nil {
			return nil, err
		}
	} else if configuration.Method == LoggingMethodHTTP {
		logExporter, err = otlploghttp.New(ctx, configuration.Http...)
		if err != nil {
			return nil, err
		}
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
		log.WithResource(res),
	)

	shutdownFunc.append(loggerProvider.ForceFlush)
	shutdownFunc.append(loggerProvider.Shutdown)
	return loggerProvider, nil
}
