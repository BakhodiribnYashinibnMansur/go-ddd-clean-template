package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
)

// ZapTracer implements tracelog.Logger interface using zap.
type ZapTracer struct {
	logger *zap.Logger
}

// NewZapTracer creates a new ZapTracer instance.
func NewZapTracer(logger *zap.Logger) *ZapTracer {
	return &ZapTracer{
		logger: logger,
	}
}

// Log implements tracelog.Logger interface.
func (t *ZapTracer) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	fields := make([]zap.Field, 0, len(data))
	for k, v := range data {
		fields = append(fields, zap.Any(k, v))
	}

	logger := t.logger.With(fields...)

	switch level {
	case tracelog.LogLevelTrace, tracelog.LogLevelDebug:
		logger.Debug(msg)
	case tracelog.LogLevelInfo:
		logger.Info(msg)
	case tracelog.LogLevelWarn:
		logger.Warn(msg)
	case tracelog.LogLevelError:
		logger.Error(msg)
	case tracelog.LogLevelNone:
		// No logging for LogLevelNone
	default:
		logger.Info(msg)
	}
}
