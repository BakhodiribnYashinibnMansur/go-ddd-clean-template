package latency

import (
	"sync"
	"time"

	hdrhistogram "github.com/HdrHistogram/hdrhistogram-go"
)

// LatencyStats holds computed percentile statistics.
type LatencyStats struct {
	P50   time.Duration `json:"p50"`
	P95   time.Duration `json:"p95"`
	P99   time.Duration `json:"p99"`
	Count int64         `json:"count"`
	Mean  time.Duration `json:"mean"`
}

// Tracker records request latencies using an HdrHistogram.
type Tracker struct {
	mu        sync.Mutex
	hist      *hdrhistogram.Histogram
	windowSec int
}

// NewTracker creates a new latency tracker.
// The histogram range is 1 microsecond to 10 seconds with 3 significant figures.
func NewTracker(windowSec int) *Tracker {
	return &Tracker{
		hist:      hdrhistogram.New(1, 10_000_000, 3), // 1µs to 10s in microseconds
		windowSec: windowSec,
	}
}

// Record adds a latency observation.
func (t *Tracker) Record(d time.Duration) {
	us := d.Microseconds()
	if us < 1 {
		us = 1
	}
	if us > 10_000_000 {
		us = 10_000_000
	}
	t.mu.Lock()
	_ = t.hist.RecordValue(us)
	t.mu.Unlock()
}

// Stats returns the current percentile statistics.
func (t *Tracker) Stats() LatencyStats {
	t.mu.Lock()
	defer t.mu.Unlock()

	count := t.hist.TotalCount()
	if count == 0 {
		return LatencyStats{}
	}
	return LatencyStats{
		P50:   time.Duration(t.hist.ValueAtQuantile(50)) * time.Microsecond,
		P95:   time.Duration(t.hist.ValueAtQuantile(95)) * time.Microsecond,
		P99:   time.Duration(t.hist.ValueAtQuantile(99)) * time.Microsecond,
		Count: count,
		Mean:  time.Duration(t.hist.Mean()) * time.Microsecond,
	}
}

// Reset clears all recorded data.
func (t *Tracker) Reset() {
	t.mu.Lock()
	t.hist.Reset()
	t.mu.Unlock()
}
