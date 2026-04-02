package latency_test

import (
	"sync"
	"testing"
	"time"

	"gct/internal/shared/infrastructure/metrics/latency"
)

func TestTracker_RecordAndStats(t *testing.T) {
	tr := latency.NewTracker(60)

	durations := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		100 * time.Millisecond,
		200 * time.Millisecond,
		500 * time.Millisecond,
	}
	for _, d := range durations {
		tr.Record(d)
	}

	stats := tr.Stats()
	if stats.Count != 6 {
		t.Errorf("expected Count=6, got %d", stats.Count)
	}
	if stats.P50 < 20*time.Millisecond || stats.P50 > 100*time.Millisecond {
		t.Errorf("P50 out of expected range: %v", stats.P50)
	}
	if stats.P95 < 200*time.Millisecond {
		t.Errorf("P95 too low: %v", stats.P95)
	}
	if stats.P99 < 200*time.Millisecond {
		t.Errorf("P99 too low: %v", stats.P99)
	}
	if stats.Mean == 0 {
		t.Error("expected Mean > 0")
	}
}

func TestTracker_Reset(t *testing.T) {
	tr := latency.NewTracker(60)
	tr.Record(50 * time.Millisecond)
	tr.Reset()
	stats := tr.Stats()
	if stats.Count != 0 {
		t.Errorf("expected Count=0 after reset, got %d", stats.Count)
	}
}

func TestTracker_EmptyStats(t *testing.T) {
	tr := latency.NewTracker(60)
	stats := tr.Stats()
	if stats.Count != 0 {
		t.Errorf("expected Count=0, got %d", stats.Count)
	}
	if stats.P50 != 0 || stats.P95 != 0 || stats.P99 != 0 || stats.Mean != 0 {
		t.Error("expected all zeroes for empty stats")
	}
}

func TestTracker_ConcurrentAccess(t *testing.T) {
	tr := latency.NewTracker(60)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tr.Record(time.Duration(i+1) * time.Millisecond)
			_ = tr.Stats()
		}(i)
	}
	wg.Wait()
	stats := tr.Stats()
	if stats.Count != 100 {
		t.Errorf("expected Count=100, got %d", stats.Count)
	}
}
