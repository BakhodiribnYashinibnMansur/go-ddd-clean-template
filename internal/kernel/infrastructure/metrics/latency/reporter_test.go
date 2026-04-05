package latency_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"gct/internal/kernel/infrastructure/metrics/latency"
)

// mockLog is a minimal mock implementing logger.Log for testing.
type mockLog struct {
	mu       sync.Mutex
	messages []string
}

func (m *mockLog) record(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
}

func (m *mockLog) count(msg string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, v := range m.messages {
		if v == msg {
			n++
		}
	}
	return n
}

func (m *mockLog) Debug(_ ...any)                                {}
func (m *mockLog) Debugf(_ string, _ ...any)                     {}
func (m *mockLog) Debugw(msg string, _ ...any)                   { m.record(msg) }
func (m *mockLog) Info(_ ...any)                                 {}
func (m *mockLog) Infof(_ string, _ ...any)                      {}
func (m *mockLog) Infow(msg string, _ ...any)                    { m.record(msg) }
func (m *mockLog) Warn(_ ...any)                                 {}
func (m *mockLog) Warnf(_ string, _ ...any)                      {}
func (m *mockLog) Warnw(msg string, _ ...any)                    { m.record(msg) }
func (m *mockLog) Error(_ ...any)                                {}
func (m *mockLog) Errorf(_ string, _ ...any)                     {}
func (m *mockLog) Errorw(msg string, _ ...any)                   { m.record(msg) }
func (m *mockLog) Fatal(_ ...any)                                {}
func (m *mockLog) Fatalf(_ string, _ ...any)                     {}
func (m *mockLog) Fatalw(msg string, _ ...any)                   { m.record(msg) }
func (m *mockLog) Debugc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLog) Infoc(_ context.Context, _ string, _ ...any)   {}
func (m *mockLog) Warnc(_ context.Context, _ string, _ ...any)   {}
func (m *mockLog) Errorc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLog) Fatalc(_ context.Context, _ string, _ ...any)  {}

func TestReporter_StartsAndStops(t *testing.T) {
	tr := latency.NewTracker(60)
	ml := &mockLog{}

	// Record some data so stats are non-zero.
	for i := 0; i < 50; i++ {
		tr.Record(time.Duration(i+1) * time.Millisecond)
	}

	r := latency.NewReporter(tr, nil, 50*time.Millisecond, 600, ml)
	r.Start(context.Background())

	time.Sleep(200 * time.Millisecond)
	r.Stop()

	if ml.count("latency stats") == 0 {
		t.Error("expected at least one 'latency stats' log entry, got none")
	}
}

func TestReporter_ResetsAfterWindow(t *testing.T) {
	tr := latency.NewTracker(60)
	ml := &mockLog{}

	for i := 0; i < 20; i++ {
		tr.Record(5 * time.Millisecond)
	}

	// interval=50ms, window=1s (windowSec=1 means window resets every 1s is too slow)
	// We use a small windowSec trick: the Reporter multiplies windowSec by time.Second,
	// but the minimum is 1 second. We set interval to 30ms and window to 1 second,
	// but sleep long enough.
	// Actually, let's just verify with a direct approach: use windowSec=1 and sleep >1s.
	// That would be slow, so instead we test at a shorter timescale.
	// The Reporter takes windowSec int, so the minimum window is 1 second.
	// For a fast test, we'll accept sleeping ~1.1s.

	r := latency.NewReporter(tr, nil, 50*time.Millisecond, 1, ml)
	r.Start(context.Background())

	// Wait for window reset (1 second + some buffer).
	time.Sleep(1200 * time.Millisecond)
	r.Stop()

	stats := tr.Stats()
	if stats.Count != 0 {
		t.Errorf("expected Count=0 after window reset, got %d", stats.Count)
	}

	if ml.count("latency tracker window reset") == 0 {
		t.Error("expected 'latency tracker window reset' log entry")
	}
}
