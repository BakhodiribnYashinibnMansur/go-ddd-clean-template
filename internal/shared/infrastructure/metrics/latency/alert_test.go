package latency_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"gct/internal/shared/infrastructure/metrics/latency"
)

type mockEnqueuer struct {
	mu    sync.Mutex
	tasks []string
}

func (m *mockEnqueuer) EnqueueTask(_ context.Context, taskType string, payload any, _ ...any) (any, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, _ := json.Marshal(payload)
	m.tasks = append(m.tasks, taskType+":"+string(b))
	return nil, nil
}

func (m *mockEnqueuer) taskCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.tasks)
}

func TestAlertManager_P95Breach(t *testing.T) {
	enq := &mockEnqueuer{}
	am := latency.NewAlertManager(enq, latency.AlertConfig{
		P95Threshold: 100 * time.Millisecond,
		P99Threshold: 500 * time.Millisecond,
		Cooldown:     time.Millisecond, // very short for test
	})

	am.Check(latency.LatencyStats{
		P50:   50 * time.Millisecond,
		P95:   200 * time.Millisecond, // breach
		P99:   300 * time.Millisecond, // no breach
		Count: 100,
		Mean:  80 * time.Millisecond,
	})

	if enq.taskCount() != 1 {
		t.Errorf("expected 1 alert, got %d", enq.taskCount())
	}
}

func TestAlertManager_P99Breach(t *testing.T) {
	enq := &mockEnqueuer{}
	am := latency.NewAlertManager(enq, latency.AlertConfig{
		P95Threshold: 200 * time.Millisecond,
		P99Threshold: 500 * time.Millisecond,
		Cooldown:     time.Millisecond,
	})

	am.Check(latency.LatencyStats{
		P50:   50 * time.Millisecond,
		P95:   100 * time.Millisecond, // no breach
		P99:   600 * time.Millisecond, // breach
		Count: 100,
		Mean:  80 * time.Millisecond,
	})

	if enq.taskCount() != 1 {
		t.Errorf("expected 1 alert, got %d", enq.taskCount())
	}
}

func TestAlertManager_NoBreach(t *testing.T) {
	enq := &mockEnqueuer{}
	am := latency.NewAlertManager(enq, latency.AlertConfig{
		P95Threshold: 200 * time.Millisecond,
		P99Threshold: 500 * time.Millisecond,
		Cooldown:     time.Millisecond,
	})

	am.Check(latency.LatencyStats{
		P50:   50 * time.Millisecond,
		P95:   100 * time.Millisecond,
		P99:   300 * time.Millisecond,
		Count: 100,
		Mean:  80 * time.Millisecond,
	})

	if enq.taskCount() != 0 {
		t.Errorf("expected 0 alerts, got %d", enq.taskCount())
	}
}

func TestAlertManager_Cooldown(t *testing.T) {
	enq := &mockEnqueuer{}
	am := latency.NewAlertManager(enq, latency.AlertConfig{
		P95Threshold: 100 * time.Millisecond,
		P99Threshold: 500 * time.Millisecond,
		Cooldown:     time.Hour, // very long cooldown
	})

	stats := latency.LatencyStats{
		P50: 50 * time.Millisecond, P95: 200 * time.Millisecond,
		P99: 300 * time.Millisecond, Count: 100, Mean: 80 * time.Millisecond,
	}

	am.Check(stats) // first: should send
	am.Check(stats) // second: should be cooled down

	if enq.taskCount() != 1 {
		t.Errorf("expected 1 alert (second should be cooled down), got %d", enq.taskCount())
	}
}

func TestAlertManager_NilEnqueuer(t *testing.T) {
	am := latency.NewAlertManager(nil, latency.AlertConfig{
		P95Threshold: 100 * time.Millisecond,
		Cooldown:     time.Millisecond,
	})

	// Should not panic
	am.Check(latency.LatencyStats{
		P95: 200 * time.Millisecond, Count: 10,
	})
}
