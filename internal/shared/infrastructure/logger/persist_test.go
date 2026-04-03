package logger

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap/zapcore"
)

// ── fieldToString ─────────────────────────────────────────────────────────────

func TestFieldToString(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"string input", "hello", "hello"},
		{"error input", errors.New("something broke"), "something broke"},
		{"int input", 42, "42"},
		{"map input", map[string]int{"a": 1}, `{"a":1}`},
		{"nil input", nil, "null"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := fieldToString(tc.input)
			if got != tc.want {
				t.Errorf("fieldToString(%v) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ── nullIfEmpty ───────────────────────────────────────────────────────────────

func TestNullIfEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input string
		isNil bool
		want  string
	}{
		{"empty string", "", true, ""},
		{"non-empty string", "foo", false, "foo"},
		{"whitespace only", "  ", false, "  "},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := nullIfEmpty(tc.input)
			if tc.isNil {
				if got != nil {
					t.Errorf("nullIfEmpty(%q) = %v, want nil", tc.input, got)
				}
			} else {
				s, ok := got.(string)
				if !ok {
					t.Fatalf("nullIfEmpty(%q) returned %T, want string", tc.input, got)
				}
				if s != tc.want {
					t.Errorf("nullIfEmpty(%q) = %q, want %q", tc.input, s, tc.want)
				}
			}
		})
	}
}

// ── parsePersistLevel ─────────────────────────────────────────────────────────

func TestParsePersistLevel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  zapcore.Level
	}{
		{"debug", "debug", zapcore.DebugLevel},
		{"info", "info", zapcore.InfoLevel},
		{"warn", "warn", zapcore.WarnLevel},
		{"error", "error", zapcore.ErrorLevel},
		{"unknown defaults to warn", "unknown", zapcore.WarnLevel},
		{"empty defaults to warn", "", zapcore.WarnLevel},
		{"case insensitive", "WARN", zapcore.WarnLevel},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parsePersistLevel(tc.input)
			if got != tc.want {
				t.Errorf("parsePersistLevel(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

// ── RedisSink.Enabled ─────────────────────────────────────────────────────────

func TestRedisSink_Enabled(t *testing.T) {
	sink := &RedisSink{minLevel: zapcore.WarnLevel}

	tests := []struct {
		level zapcore.Level
		want  bool
	}{
		{zapcore.DebugLevel, false},
		{zapcore.InfoLevel, false},
		{zapcore.WarnLevel, true},
		{zapcore.ErrorLevel, true},
	}

	for _, tc := range tests {
		t.Run(tc.level.String(), func(t *testing.T) {
			got := sink.Enabled(tc.level)
			if got != tc.want {
				t.Errorf("Enabled(%v) = %v, want %v", tc.level, got, tc.want)
			}
		})
	}
}

// ── RedisSink.With ────────────────────────────────────────────────────────────

func TestRedisSink_With(t *testing.T) {
	sink := &RedisSink{minLevel: zapcore.WarnLevel}
	got := sink.With(nil)
	if got != sink {
		t.Error("With() should return the same instance")
	}
}

// ── RedisSink.Sync ────────────────────────────────────────────────────────────

func TestRedisSink_Sync(t *testing.T) {
	sink := &RedisSink{}
	if err := sink.Sync(); err != nil {
		t.Errorf("Sync() = %v, want nil", err)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newTestSink(t *testing.T) (*RedisSink, *miniredis.Miniredis) {
	t.Helper()
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	t.Cleanup(func() { rdb.Close() })

	sink := &RedisSink{
		rdb:         rdb,
		key:         "test:logs",
		minLevel:    zapcore.WarnLevel,
		pushTimeout: time.Second,
	}
	return sink, s
}

func makeEntry(level zapcore.Level, msg string) zapcore.Entry {
	return zapcore.Entry{
		Level:   level,
		Message: msg,
		Time:    time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
	}
}

// ── RedisSink.Write ───────────────────────────────────────────────────────────

func TestRedisSink_Write(t *testing.T) {
	sink, s := newTestSink(t)

	ent := makeEntry(zapcore.ErrorLevel, "something failed")
	if err := sink.Write(ent, nil); err != nil {
		t.Fatalf("Write() returned error: %v", err)
	}

	items, err := s.List(sink.key)
	if err != nil {
		t.Fatalf("failed to read Redis list: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item in Redis list, got %d", len(items))
	}

	var got logEntry
	if err := json.Unmarshal([]byte(items[0]), &got); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}
	if got.Level != "error" {
		t.Errorf("level = %q, want %q", got.Level, "error")
	}
	if got.Message != "something failed" {
		t.Errorf("message = %q, want %q", got.Message, "something failed")
	}
	if got.Timestamp == "" {
		t.Error("timestamp should not be empty")
	}
}

// ── RedisSink.Write with fields ───────────────────────────────────────────────

func TestRedisSink_Write_WithFields(t *testing.T) {
	sink, s := newTestSink(t)

	ent := makeEntry(zapcore.WarnLevel, "slow query")
	fields := []zapcore.Field{
		zapcore.Field{Key: "operation", Type: zapcore.StringType, String: "ListUsers"},
		zapcore.Field{Key: "entity", Type: zapcore.StringType, String: "user"},
	}

	if err := sink.Write(ent, fields); err != nil {
		t.Fatalf("Write() returned error: %v", err)
	}

	items, err := s.List(sink.key)
	if err != nil {
		t.Fatalf("failed to read Redis list: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	var got logEntry
	if err := json.Unmarshal([]byte(items[0]), &got); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if got.Operation != "ListUsers" {
		t.Errorf("operation = %q, want %q", got.Operation, "ListUsers")
	}
	if got.Entity != "user" {
		t.Errorf("entity = %q, want %q", got.Entity, "user")
	}
}

// ── RedisSink.Write silent fail when Redis is down ────────────────────────────

func TestRedisSink_Write_SilentFailOnRedisDown(t *testing.T) {
	sink, s := newTestSink(t)

	// Shut down Redis before writing.
	s.Close()

	ent := makeEntry(zapcore.ErrorLevel, "should not crash")
	err := sink.Write(ent, nil)
	if err != nil {
		t.Errorf("Write() should silently fail, got error: %v", err)
	}
}

// ── NewRedisSink ──────────────────────────────────────────────────────────────

func TestNewRedisSink(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	t.Cleanup(func() { rdb.Close() })

	core := NewRedisSink(rdb, PersistConfig{
		Level:    "error",
		RedisKey: "app:logs",
	})

	if core == nil {
		t.Fatal("NewRedisSink returned nil")
	}

	// The returned core should respect the configured level.
	if !core.Enabled(zapcore.ErrorLevel) {
		t.Error("core should be enabled at error level")
	}
	if core.Enabled(zapcore.WarnLevel) {
		t.Error("core should not be enabled at warn level when configured for error")
	}
}
