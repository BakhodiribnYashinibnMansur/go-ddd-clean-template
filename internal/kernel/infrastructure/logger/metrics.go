package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap/zapcore"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	logCounter  metric.Int64Counter
	metricsInit bool
)

func initMetrics() {
	if metricsInit {
		return
	}
	meter := otel.Meter("logger")
	var err error
	logCounter, err = meter.Int64Counter("app.log.entries",
		metric.WithDescription("Total log entries by level and operation"),
		metric.WithUnit("{entry}"),
	)
	if err != nil {
		return
	}
	metricsInit = true
}

// metricsCore wraps a zapcore.Core and increments OTel counters per log entry.
type metricsCore struct {
	zapcore.Core
}

// NewMetricsCore wraps a core to count log entries by level via OTel metrics.
func NewMetricsCore(core zapcore.Core) zapcore.Core {
	initMetrics()
	if !metricsInit {
		return core // metrics not available, skip
	}
	return &metricsCore{Core: core}
}

func (c *metricsCore) With(fields []zapcore.Field) zapcore.Core {
	return &metricsCore{Core: c.Core.With(fields)}
}

func (c *metricsCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Core.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *metricsCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	// Called from the synchronous zapcore Write path. The logging core has no
	// caller context and the metric increment must not inherit any request
	// cancellation that could drop the counter update.
	if logCounter != nil {
		attrs := []attribute.KeyValue{
			attribute.String("level", ent.Level.String()),
		}
		// Extract operation from fields if present
		for _, f := range fields {
			if f.Key == "operation" && f.Type == zapcore.StringType {
				attrs = append(attrs, attribute.String("operation", f.String))
				break
			}
		}
		logCounter.Add(context.Background(), 1, metric.WithAttributes(attrs...))
	}
	if err := c.Core.Write(ent, fields); err != nil {
		return fmt.Errorf("logger.metricsCore.Write: %w", err)
	}
	return nil
}
