package logger

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
)

type loggerKeyType int

const loggerKey loggerKeyType = 0

func Set(ctx context.Context, lgr log.LoggerProvider) context.Context {
	return context.WithValue(ctx, loggerKey, lgr)
}

func Get(ctx context.Context) log.LoggerProvider {
	value := ctx.Value(loggerKey)
	if value == nil {
		return global.GetLoggerProvider()
	}
	lgr, ok := value.(log.LoggerProvider)
	if !ok {
		return global.GetLoggerProvider()
	}
	return lgr
}

func Wrap(logProvider log.LoggerProvider, name string) *slog.Logger {
	return slog.New(
		otelslog.NewHandler(
			name,
			otelslog.WithLoggerProvider(logProvider),
		),
	)
}
