package seeder

import (
	"context"
	"testing"

	"gct/config"
)

// ---------------------------------------------------------------------------
// Mock logger
// ---------------------------------------------------------------------------

type testLogger struct{}

func (l *testLogger) Debug(args ...any)                                    {}
func (l *testLogger) Debugf(template string, args ...any)                  {}
func (l *testLogger) Debugw(msg string, keysAndValues ...any)              {}
func (l *testLogger) Info(args ...any)                                     {}
func (l *testLogger) Infof(template string, args ...any)                   {}
func (l *testLogger) Infow(msg string, keysAndValues ...any)               {}
func (l *testLogger) Warn(args ...any)                                     {}
func (l *testLogger) Warnf(template string, args ...any)                   {}
func (l *testLogger) Warnw(msg string, keysAndValues ...any)               {}
func (l *testLogger) Error(args ...any)                                    {}
func (l *testLogger) Errorf(template string, args ...any)                  {}
func (l *testLogger) Errorw(msg string, keysAndValues ...any)              {}
func (l *testLogger) Fatal(args ...any)                                    {}
func (l *testLogger) Fatalf(template string, args ...any)                  {}
func (l *testLogger) Fatalw(msg string, keysAndValues ...any)              {}
func (l *testLogger) Debugc(_ context.Context, _ string, _ ...any)         {}
func (l *testLogger) Infoc(_ context.Context, _ string, _ ...any)          {}
func (l *testLogger) Warnc(_ context.Context, _ string, _ ...any)          {}
func (l *testLogger) Errorc(_ context.Context, _ string, _ ...any)         {}
func (l *testLogger) Fatalc(_ context.Context, _ string, _ ...any)         {}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestNew(t *testing.T) {
	cfg := &config.Config{}
	s := New(nil, &testLogger{}, cfg)

	if s == nil {
		t.Fatal("expected non-nil seeder")
	}
	if s.cfg != cfg {
		t.Fatal("config mismatch")
	}
	if s.pool != nil {
		t.Fatal("expected nil pool")
	}
}

func TestSeed_DisabledSeeder_NilCounts(t *testing.T) {
	cfg := &config.Config{}
	cfg.Seeder.Enabled = false
	s := New(nil, &testLogger{}, cfg)

	err := s.Seed(context.Background(), nil)
	if err != nil {
		t.Fatalf("expected nil error for disabled seeder, got %v", err)
	}
}

func TestGetCount_WithCustomCount(t *testing.T) {
	cfg := &config.Config{}
	s := New(nil, &testLogger{}, cfg)

	counts := map[string]int{"users": 42}
	result := s.getCount(counts, "users", 10)
	if result != 42 {
		t.Fatalf("expected 42, got %d", result)
	}
}

func TestGetCount_WithoutCustomCount(t *testing.T) {
	cfg := &config.Config{}
	s := New(nil, &testLogger{}, cfg)

	counts := map[string]int{"roles": 5}
	result := s.getCount(counts, "users", 10)
	if result != 10 {
		t.Fatalf("expected default 10, got %d", result)
	}
}

func TestGetCount_NilCounts(t *testing.T) {
	cfg := &config.Config{}
	s := New(nil, &testLogger{}, cfg)

	result := s.getCount(nil, "users", 25)
	if result != 25 {
		t.Fatalf("expected default 25, got %d", result)
	}
}

func TestGetCount_ZeroCustomCount(t *testing.T) {
	cfg := &config.Config{}
	s := New(nil, &testLogger{}, cfg)

	counts := map[string]int{"users": 0}
	result := s.getCount(counts, "users", 10)
	if result != 0 {
		t.Fatalf("expected 0 (explicit custom), got %d", result)
	}
}
