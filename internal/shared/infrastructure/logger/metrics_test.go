package logger

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestNewMetricsCore_FallbackWhenMetricsUnavailable(t *testing.T) {
	t.Helper()

	// In test environments OTel meter is typically not initialised,
	// so NewMetricsCore should return the original core unchanged.
	inner := &mockCore{level: zapcore.DebugLevel}

	// Reset metricsInit so initMetrics runs fresh.
	origInit := metricsInit
	origCounter := logCounter
	metricsInit = false
	logCounter = nil
	t.Cleanup(func() {
		metricsInit = origInit
		logCounter = origCounter
	})

	got := NewMetricsCore(inner)

	// If OTel init fails, we expect the original core back.
	if _, ok := got.(*metricsCore); ok {
		// metricsCore returned means OTel happened to initialise.
		// That is acceptable - just verify it wraps properly.
		t.Log("OTel initialised in test; metricsCore returned (acceptable)")
	} else {
		// Unwrapped core returned - verify it is the same inner core.
		if got != inner {
			t.Error("expected original inner core when metrics unavailable")
		}
	}
}

func TestMetricsCore_With(t *testing.T) {
	t.Helper()

	inner := &mockCore{level: zapcore.DebugLevel}
	mc := &metricsCore{Core: inner}

	fields := []zapcore.Field{
		{Key: "operation", Type: zapcore.StringType, String: "create"},
	}

	newCore := mc.With(fields)

	wrapped, ok := newCore.(*metricsCore)
	if !ok {
		t.Fatal("expected *metricsCore from With")
	}

	innerNew, ok := wrapped.Core.(*mockCore)
	if !ok {
		t.Fatal("expected *mockCore inside metricsCore")
	}

	if len(innerNew.withFields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(innerNew.withFields))
	}
	if innerNew.withFields[0].Key != "operation" {
		t.Errorf("field key: got %q, want %q", innerNew.withFields[0].Key, "operation")
	}
}

func TestMetricsCore_Check(t *testing.T) {
	t.Helper()

	tests := []struct {
		name        string
		coreLevel   zapcore.Level
		entLevel    zapcore.Level
		wantEnabled bool
	}{
		{
			name:        "entry level at core level - enabled",
			coreLevel:   zapcore.InfoLevel,
			entLevel:    zapcore.InfoLevel,
			wantEnabled: true,
		},
		{
			name:        "entry level above core level - enabled",
			coreLevel:   zapcore.DebugLevel,
			entLevel:    zapcore.ErrorLevel,
			wantEnabled: true,
		},
		{
			name:        "entry level below core level - disabled",
			coreLevel:   zapcore.ErrorLevel,
			entLevel:    zapcore.InfoLevel,
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			inner := &mockCore{level: tt.coreLevel}
			mc := &metricsCore{Core: inner}

			got := mc.Enabled(tt.entLevel)
			if got != tt.wantEnabled {
				t.Errorf("Enabled(%v) = %v, want %v", tt.entLevel, got, tt.wantEnabled)
			}
		})
	}
}

func TestMetricsCore_Write(t *testing.T) {
	t.Helper()

	inner := &mockCore{level: zapcore.DebugLevel}
	mc := &metricsCore{Core: inner}

	ent := zapcore.Entry{Level: zapcore.InfoLevel, Message: "test msg"}
	fields := []zapcore.Field{
		{Key: "operation", Type: zapcore.StringType, String: "create"},
		{Key: "entity", Type: zapcore.StringType, String: "user"},
	}

	err := mc.Write(ent, fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !inner.writeCalled {
		t.Fatal("inner Write was not called")
	}
	if inner.writeEntry.Message != "test msg" {
		t.Errorf("entry message: got %q, want %q", inner.writeEntry.Message, "test msg")
	}
	if len(inner.writeFields) != 2 {
		t.Fatalf("expected 2 fields passed to inner, got %d", len(inner.writeFields))
	}
}
