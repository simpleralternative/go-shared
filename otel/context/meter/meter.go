package meter

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type meterKeyType int

const meterKey meterKeyType = 0

func Set(ctx context.Context, mtr metric.MeterProvider) context.Context {
	return context.WithValue(ctx, meterKey, mtr)
}

func Get(ctx context.Context) metric.MeterProvider {
	value := ctx.Value(meterKey)
	if value == nil {
		return otel.GetMeterProvider()
	}
	mtr, ok := value.(metric.MeterProvider)
	if !ok {
		return otel.GetMeterProvider()
	}
	return mtr
}
