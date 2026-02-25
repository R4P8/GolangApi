package helpers

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type HTTPMetrics struct {
	Throughput  metric.Int64Counter
	ErrorCount  metric.Int64Counter
	LatencyHist metric.Float64Histogram
}

func Observe(
	ctx context.Context,
	metrics HTTPMetrics,
	operation string,
	start time.Time,
	err error,
) {
	duration := float64(time.Since(start).Milliseconds())

	metrics.Throughput.Add(ctx, 1,
		metric.WithAttributes(attribute.String("operation", operation)),
	)

	metrics.LatencyHist.Record(ctx, duration,
		metric.WithAttributes(attribute.String("operation", operation)),
	)

	if err != nil {
		metrics.ErrorCount.Add(ctx, 1,
			metric.WithAttributes(attribute.String("operation", operation)),
		)
	}
}
