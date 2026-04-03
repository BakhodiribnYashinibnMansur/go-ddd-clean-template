package logger

import (
	"context"
	"sync"
	"testing"
	"time"
)

// mockLog implements the Log interface for testing.
// Only Warnc records entries; all other methods are no-ops.
type mockLog struct {
	mu           sync.Mutex
	warnMessages []mockLogEntry
}

type mockLogEntry struct {
	msg           string
	keysAndValues []any
}

func (m *mockLog) recordWarn(msg string, kv []any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warnMessages = append(m.warnMessages, mockLogEntry{msg: msg, keysAndValues: kv})
}

func (m *mockLog) getWarnMessages() []mockLogEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]mockLogEntry, len(m.warnMessages))
	copy(cp, m.warnMessages)
	return cp
}

func (m *mockLog) Debug(_ ...any)                                          {}
func (m *mockLog) Debugf(_ string, _ ...any)                               {}
func (m *mockLog) Debugw(_ string, _ ...any)                               {}
func (m *mockLog) Info(_ ...any)                                           {}
func (m *mockLog) Infof(_ string, _ ...any)                                {}
func (m *mockLog) Infow(_ string, _ ...any)                                {}
func (m *mockLog) Warn(_ ...any)                                           {}
func (m *mockLog) Warnf(_ string, _ ...any)                                {}
func (m *mockLog) Warnw(_ string, _ ...any)                                {}
func (m *mockLog) Error(_ ...any)                                          {}
func (m *mockLog) Errorf(_ string, _ ...any)                               {}
func (m *mockLog) Errorw(_ string, _ ...any)                               {}
func (m *mockLog) Fatal(_ ...any)                                          {}
func (m *mockLog) Fatalf(_ string, _ ...any)                               {}
func (m *mockLog) Fatalw(_ string, _ ...any)                               {}
func (m *mockLog) Debugc(_ context.Context, _ string, _ ...any)            {}
func (m *mockLog) Infoc(_ context.Context, _ string, _ ...any)             {}
func (m *mockLog) Errorc(_ context.Context, _ string, _ ...any)            {}
func (m *mockLog) Fatalc(_ context.Context, _ string, _ ...any)            {}
func (m *mockLog) Warnc(_ context.Context, msg string, kv ...any) {
	m.recordWarn(msg, kv)
}

func TestSlowOp(t *testing.T) {
	t.Helper()

	// Save and restore original threshold after tests.
	origThreshold := slowOpThreshold
	t.Cleanup(func() { slowOpThreshold = origThreshold })

	tests := []struct {
		name        string
		threshold   time.Duration
		sleepFor    time.Duration
		wantWarning bool
	}{
		{
			name:        "under threshold - no warning",
			threshold:   200 * time.Millisecond,
			sleepFor:    10 * time.Millisecond,
			wantWarning: false,
		},
		{
			name:        "over threshold - warning logged",
			threshold:   10 * time.Millisecond,
			sleepFor:    30 * time.Millisecond,
			wantWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			slowOpThreshold = tt.threshold
			ml := &mockLog{}
			ctx := context.Background()

			done := SlowOp(ml, ctx, "TestOp", "TestEntity")
			time.Sleep(tt.sleepFor)
			done()

			msgs := ml.getWarnMessages()
			if tt.wantWarning && len(msgs) == 0 {
				t.Fatal("expected a warning but none was logged")
			}
			if !tt.wantWarning && len(msgs) != 0 {
				t.Fatalf("expected no warning but got %d", len(msgs))
			}

			if tt.wantWarning {
				entry := msgs[0]
				if entry.msg != "slow operation" {
					t.Errorf("got msg %q, want %q", entry.msg, "slow operation")
				}
				assertKV(t, entry.keysAndValues, "operation", "TestOp")
				assertKV(t, entry.keysAndValues, "entity", "TestEntity")
				assertKVExists(t, entry.keysAndValues, "duration_ms")
				assertKVExists(t, entry.keysAndValues, "threshold_ms")
			}
		})
	}
}

func TestSetSlowOpThreshold(t *testing.T) {
	t.Helper()

	origThreshold := slowOpThreshold
	t.Cleanup(func() { slowOpThreshold = origThreshold })

	tests := []struct {
		name          string
		initial       time.Duration
		set           time.Duration
		wantThreshold time.Duration
	}{
		{
			name:          "valid positive duration changes threshold",
			initial:       500 * time.Millisecond,
			set:           1 * time.Second,
			wantThreshold: 1 * time.Second,
		},
		{
			name:          "zero duration does not change threshold",
			initial:       500 * time.Millisecond,
			set:           0,
			wantThreshold: 500 * time.Millisecond,
		},
		{
			name:          "negative duration does not change threshold",
			initial:       500 * time.Millisecond,
			set:           -1 * time.Second,
			wantThreshold: 500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			slowOpThreshold = tt.initial
			SetSlowOpThreshold(tt.set)

			if slowOpThreshold != tt.wantThreshold {
				t.Errorf("threshold = %v, want %v", slowOpThreshold, tt.wantThreshold)
			}
		})
	}
}

func TestSlowOp_Concurrent(t *testing.T) {
	t.Helper()

	origThreshold := slowOpThreshold
	t.Cleanup(func() { slowOpThreshold = origThreshold })

	slowOpThreshold = 5 * time.Millisecond
	ml := &mockLog{}
	ctx := context.Background()

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			done := SlowOp(ml, ctx, "ConcurrentOp", "Entity")
			time.Sleep(20 * time.Millisecond)
			done()
		}()
	}

	wg.Wait()

	msgs := ml.getWarnMessages()
	if len(msgs) != goroutines {
		t.Errorf("expected %d warnings, got %d", goroutines, len(msgs))
	}
}

// assertKV checks that a key exists in the keysAndValues slice with the expected value.
func assertKV(t *testing.T, kv []any, key string, want any) {
	t.Helper()
	for i := 0; i < len(kv)-1; i += 2 {
		if kv[i] == key {
			if kv[i+1] != want {
				t.Errorf("key %q: got %v, want %v", key, kv[i+1], want)
			}
			return
		}
	}
	t.Errorf("key %q not found in keysAndValues", key)
}

// assertKVExists checks that a key exists in the keysAndValues slice.
func assertKVExists(t *testing.T, kv []any, key string) {
	t.Helper()
	for i := 0; i < len(kv)-1; i += 2 {
		if kv[i] == key {
			return
		}
	}
	t.Errorf("key %q not found in keysAndValues", key)
}
