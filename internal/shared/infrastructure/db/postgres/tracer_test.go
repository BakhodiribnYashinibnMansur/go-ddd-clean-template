package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
)

func TestNewZapTracer(t *testing.T) {
	tracer := NewZapTracer(zap.NewNop())
	if tracer == nil {
		t.Fatal("expected non-nil ZapTracer")
	}
}

func TestZapTracer_Log(t *testing.T) {
	tracer := NewZapTracer(zap.NewNop())
	ctx := context.Background()

	levels := []struct {
		name  string
		level tracelog.LogLevel
	}{
		{"Trace", tracelog.LogLevelTrace},
		{"Debug", tracelog.LogLevelDebug},
		{"Info", tracelog.LogLevelInfo},
		{"Warn", tracelog.LogLevelWarn},
		{"Error", tracelog.LogLevelError},
		{"None", tracelog.LogLevelNone},
		{"unknown", tracelog.LogLevel(99)},
	}

	for _, tt := range levels {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic.
			tracer.Log(ctx, tt.level, "test message", nil)
		})
	}
}

func TestZapTracer_Log_WithData(t *testing.T) {
	tracer := NewZapTracer(zap.NewNop())
	ctx := context.Background()

	data := map[string]any{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	// Should not panic when data fields are provided.
	tracer.Log(ctx, tracelog.LogLevelInfo, "test with data", data)
}
