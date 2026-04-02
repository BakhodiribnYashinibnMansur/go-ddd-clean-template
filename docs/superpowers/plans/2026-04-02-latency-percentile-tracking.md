# Latency Percentile Tracking (p95/p99) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Real-time p95/p99 latency tracking with in-memory HDR Histogram, periodic logging, Telegram alerting, REST endpoint, and Grafana dashboard.

**Architecture:** HDR Histogram records every request latency in O(1) time/space. A periodic reporter goroutine reads percentile stats, logs them, and triggers Telegram alerts via Asynq when thresholds are breached. A `/metrics/latency` endpoint exposes real-time stats as JSON. Grafana dashboard provides PromQL-based visualization.

**Tech Stack:** `github.com/HdrHistogram/hdrhistogram-go`, Gin middleware, Asynq (Telegram), OTel/Prometheus (existing), Grafana

---

## File Structure

```
internal/shared/infrastructure/metrics/latency/
├── tracker.go       # HDR Histogram wrapper — Record, Stats, Reset (thread-safe)
├── tracker_test.go  # Unit tests for tracker
├── alert.go         # Threshold checking + Asynq→Telegram alert with cooldown
├── alert_test.go    # Unit tests for alert manager
├── reporter.go      # Periodic goroutine: Stats→log→alert
├── reporter_test.go # Unit tests for reporter

internal/shared/infrastructure/middleware/
├── latency_tracker.go      # Gin middleware — records latency to tracker
├── latency_tracker_test.go # Unit tests for middleware

docs/grafana/
├── latency-dashboard.json  # Grafana dashboard import

Modify:
├── config/app_settings.go          # Add latency config fields to Metrics struct
├── config/config.yaml              # Add latency config defaults
├── internal/shared/infrastructure/middleware/setup.go  # Register latency middleware
├── internal/app/app.go             # Initialize tracker, alert, reporter
├── internal/app/infra_routes.go    # Add /metrics/latency endpoint
```

---

### Task 1: Add HDR Histogram Dependency

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: Install hdrhistogram-go**

```bash
cd "/Users/mrb/Desktop/Golang Template/Backend"
go get github.com/HdrHistogram/hdrhistogram-go@latest
```

- [ ] **Step 2: Verify installation**

```bash
go mod tidy
grep hdrhistogram go.mod
```

Expected: line containing `github.com/HdrHistogram/hdrhistogram-go`

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add hdrhistogram-go for latency percentile tracking"
```

---

### Task 2: Add Config Fields

**Files:**
- Modify: `config/app_settings.go:44-47`
- Modify: `config/config.yaml:22-24`

- [ ] **Step 1: Write config test**

Create `config/app_settings_latency_test.go`:

```go
package config

import (
	"testing"
)

func TestMetricsLatencyDefaults(t *testing.T) {
	m := Metrics{
		Enabled:               true,
		LatencyEnabled:        true,
		LatencyP95Threshold:   "200ms",
		LatencyP99Threshold:   "500ms",
		LatencyWindowSec:      300,
		LatencyLogIntervalSec: 60,
	}

	if !m.LatencyEnabled {
		t.Error("LatencyEnabled should be true")
	}
	if m.LatencyP95Threshold != "200ms" {
		t.Errorf("LatencyP95Threshold = %q, want %q", m.LatencyP95Threshold, "200ms")
	}
	if m.LatencyP99Threshold != "500ms" {
		t.Errorf("LatencyP99Threshold = %q, want %q", m.LatencyP99Threshold, "500ms")
	}
	if m.LatencyWindowSec != 300 {
		t.Errorf("LatencyWindowSec = %d, want %d", m.LatencyWindowSec, 300)
	}
	if m.LatencyLogIntervalSec != 60 {
		t.Errorf("LatencyLogIntervalSec = %d, want %d", m.LatencyLogIntervalSec, 60)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./config/ -run TestMetricsLatencyDefaults -v
```

Expected: FAIL — `LatencyEnabled` field does not exist.

- [ ] **Step 3: Add fields to Metrics struct**

In `config/app_settings.go`, replace the existing `Metrics` struct:

```go
// Metrics -.
Metrics struct {
    Enabled            bool   `yaml:"enabled"`
    SlowQueryThreshold string `yaml:"slow_query_threshold" env:"METRICS_SLOW_QUERY_THRESHOLD" envDefault:"100ms"`

    // Latency percentile tracking
    LatencyEnabled        bool   `yaml:"latency_enabled" env:"METRICS_LATENCY_ENABLED" envDefault:"true"`
    LatencyP95Threshold   string `yaml:"latency_p95_threshold" env:"METRICS_LATENCY_P95_THRESHOLD" envDefault:"200ms"`
    LatencyP99Threshold   string `yaml:"latency_p99_threshold" env:"METRICS_LATENCY_P99_THRESHOLD" envDefault:"500ms"`
    LatencyWindowSec      int    `yaml:"latency_window_sec" env:"METRICS_LATENCY_WINDOW_SEC" envDefault:"300"`
    LatencyLogIntervalSec int    `yaml:"latency_log_interval_sec" env:"METRICS_LATENCY_LOG_INTERVAL_SEC" envDefault:"60"`
}
```

- [ ] **Step 4: Add config.yaml defaults**

In `config/config.yaml`, under the `metrics:` section, add:

```yaml
metrics:
  enabled: true
  slow_query_threshold: "100ms"
  latency_enabled: true
  latency_p95_threshold: "200ms"
  latency_p99_threshold: "500ms"
  latency_window_sec: 300
  latency_log_interval_sec: 60
```

- [ ] **Step 5: Run test to verify it passes**

```bash
go test ./config/ -run TestMetricsLatencyDefaults -v
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add config/app_settings.go config/config.yaml config/app_settings_latency_test.go
git commit -m "feat(config): add latency percentile tracking settings"
```

---

### Task 3: Implement LatencyTracker

**Files:**
- Create: `internal/shared/infrastructure/metrics/latency/tracker.go`
- Create: `internal/shared/infrastructure/metrics/latency/tracker_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/shared/infrastructure/metrics/latency/tracker_test.go`:

```go
package latency

import (
	"testing"
	"time"
)

func TestTracker_RecordAndStats(t *testing.T) {
	tr := NewTracker(300)

	// Record some latencies
	tr.Record(10 * time.Millisecond)
	tr.Record(20 * time.Millisecond)
	tr.Record(50 * time.Millisecond)
	tr.Record(100 * time.Millisecond)
	tr.Record(200 * time.Millisecond)

	stats := tr.Stats()

	if stats.Count != 5 {
		t.Errorf("Count = %d, want 5", stats.Count)
	}
	if stats.P50 < 20*time.Millisecond || stats.P50 > 100*time.Millisecond {
		t.Errorf("P50 = %v, expected between 20ms-100ms", stats.P50)
	}
	if stats.P95 < 100*time.Millisecond {
		t.Errorf("P95 = %v, expected >= 100ms", stats.P95)
	}
	if stats.P99 < 100*time.Millisecond {
		t.Errorf("P99 = %v, expected >= 100ms", stats.P99)
	}
	if stats.Max < 200*time.Millisecond {
		t.Errorf("Max = %v, expected >= 200ms", stats.Max)
	}
	if stats.Window != "5m0s" {
		t.Errorf("Window = %q, want %q", stats.Window, "5m0s")
	}
}

func TestTracker_Reset(t *testing.T) {
	tr := NewTracker(300)

	tr.Record(100 * time.Millisecond)
	tr.Reset()

	stats := tr.Stats()
	if stats.Count != 0 {
		t.Errorf("Count after reset = %d, want 0", stats.Count)
	}
}

func TestTracker_EmptyStats(t *testing.T) {
	tr := NewTracker(300)
	stats := tr.Stats()

	if stats.Count != 0 {
		t.Errorf("Count = %d, want 0", stats.Count)
	}
	if stats.P95 != 0 {
		t.Errorf("P95 = %v, want 0", stats.P95)
	}
}

func TestTracker_ConcurrentAccess(t *testing.T) {
	tr := NewTracker(300)

	done := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func(d int) {
			tr.Record(time.Duration(d) * time.Millisecond)
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < 100; i++ {
		<-done
	}

	stats := tr.Stats()
	if stats.Count != 100 {
		t.Errorf("Count = %d, want 100", stats.Count)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/shared/infrastructure/metrics/latency/ -v
```

Expected: FAIL — package does not exist.

- [ ] **Step 3: Implement tracker**

Create `internal/shared/infrastructure/metrics/latency/tracker.go`:

```go
package latency

import (
	"sync"
	"time"

	hdr "github.com/HdrHistogram/hdrhistogram-go"
)

// Stats holds computed percentile statistics.
type Stats struct {
	P50   time.Duration `json:"p50"`
	P95   time.Duration `json:"p95"`
	P99   time.Duration `json:"p99"`
	P999  time.Duration `json:"p999"`
	Mean  time.Duration `json:"mean"`
	Max   time.Duration `json:"max"`
	Count int64         `json:"count"`
	Window string       `json:"window"`
}

// Tracker records request latencies using an HDR Histogram.
// Thread-safe for concurrent use.
type Tracker struct {
	mu     sync.Mutex
	hist   *hdr.Histogram
	window time.Duration
}

// NewTracker creates a new latency tracker.
// windowSec defines the reporting window duration in seconds.
// Histogram range: 1µs to 10s, 3 significant figures.
func NewTracker(windowSec int) *Tracker {
	return &Tracker{
		hist:   hdr.New(1, 10_000_000, 3), // 1µs to 10s in microseconds
		window: time.Duration(windowSec) * time.Second,
	}
}

// Record records a request latency.
func (t *Tracker) Record(d time.Duration) {
	us := d.Microseconds()
	if us < 1 {
		us = 1
	}
	if us > 10_000_000 {
		us = 10_000_000
	}

	t.mu.Lock()
	t.hist.RecordValue(us)
	t.mu.Unlock()
}

// Stats returns computed percentile statistics.
func (t *Tracker) Stats() Stats {
	t.mu.Lock()
	defer t.mu.Unlock()

	count := t.hist.TotalCount()
	if count == 0 {
		return Stats{Window: t.window.String()}
	}

	return Stats{
		P50:   time.Duration(t.hist.ValueAtQuantile(50)) * time.Microsecond,
		P95:   time.Duration(t.hist.ValueAtQuantile(95)) * time.Microsecond,
		P99:   time.Duration(t.hist.ValueAtQuantile(99)) * time.Microsecond,
		P999:  time.Duration(t.hist.ValueAtQuantile(99.9)) * time.Microsecond,
		Mean:  time.Duration(t.hist.Mean()) * time.Microsecond,
		Max:   time.Duration(t.hist.Max()) * time.Microsecond,
		Count: count,
		Window: t.window.String(),
	}
}

// Reset clears all recorded data.
func (t *Tracker) Reset() {
	t.mu.Lock()
	t.hist.Reset()
	t.mu.Unlock()
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/shared/infrastructure/metrics/latency/ -v
```

Expected: All 4 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/shared/infrastructure/metrics/latency/tracker.go internal/shared/infrastructure/metrics/latency/tracker_test.go
git commit -m "feat(latency): implement HDR Histogram latency tracker"
```

---

### Task 4: Implement AlertManager

**Files:**
- Create: `internal/shared/infrastructure/metrics/latency/alert.go`
- Create: `internal/shared/infrastructure/metrics/latency/alert_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/shared/infrastructure/metrics/latency/alert_test.go`:

```go
package latency

import (
	"context"
	"sync"
	"testing"
	"time"
)

type mockEnqueuer struct {
	mu       sync.Mutex
	payloads []any
}

func (m *mockEnqueuer) EnqueueTask(_ context.Context, _ string, payload any, _ ...any) (any, error) {
	m.mu.Lock()
	m.payloads = append(m.payloads, payload)
	m.mu.Unlock()
	return nil, nil
}

func (m *mockEnqueuer) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.payloads)
}

type mockLogger struct{}

func (m *mockLogger) Info(args ...any)                                   {}
func (m *mockLogger) Infof(format string, args ...any)                   {}
func (m *mockLogger) Infow(msg string, keysAndValues ...any)             {}
func (m *mockLogger) Infoc(ctx context.Context, msg string, kv ...any)   {}
func (m *mockLogger) Debug(args ...any)                                  {}
func (m *mockLogger) Debugf(format string, args ...any)                  {}
func (m *mockLogger) Debugw(msg string, keysAndValues ...any)            {}
func (m *mockLogger) Debugc(ctx context.Context, msg string, kv ...any)  {}
func (m *mockLogger) Warn(args ...any)                                   {}
func (m *mockLogger) Warnf(format string, args ...any)                   {}
func (m *mockLogger) Warnw(msg string, keysAndValues ...any)             {}
func (m *mockLogger) Warnc(ctx context.Context, msg string, kv ...any)   {}
func (m *mockLogger) Error(args ...any)                                  {}
func (m *mockLogger) Errorf(format string, args ...any)                  {}
func (m *mockLogger) Errorw(msg string, keysAndValues ...any)            {}
func (m *mockLogger) Errorc(ctx context.Context, msg string, kv ...any)  {}
func (m *mockLogger) Fatal(args ...any)                                  {}
func (m *mockLogger) Fatalf(format string, args ...any)                  {}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any)            {}
func (m *mockLogger) Fatalc(ctx context.Context, msg string, kv ...any)  {}
func (m *mockLogger) Panic(args ...any)                                  {}
func (m *mockLogger) Panicf(format string, args ...any)                  {}

func TestAlertManager_P95Breach(t *testing.T) {
	enq := &mockEnqueuer{}
	am := NewAlertManager(AlertConfig{
		P95Threshold: 200 * time.Millisecond,
		P99Threshold: 500 * time.Millisecond,
		CooldownSec:  1,
	}, enq, &mockLogger{})

	stats := Stats{
		P95:   250 * time.Millisecond, // breaches 200ms
		P99:   400 * time.Millisecond, // under 500ms
		Count: 100,
	}

	am.Check(stats)

	if enq.count() != 1 {
		t.Errorf("expected 1 alert, got %d", enq.count())
	}
}

func TestAlertManager_P99Breach(t *testing.T) {
	enq := &mockEnqueuer{}
	am := NewAlertManager(AlertConfig{
		P95Threshold: 200 * time.Millisecond,
		P99Threshold: 500 * time.Millisecond,
		CooldownSec:  1,
	}, enq, &mockLogger{})

	stats := Stats{
		P95:   250 * time.Millisecond,
		P99:   600 * time.Millisecond, // breaches 500ms
		Count: 100,
	}

	am.Check(stats)

	// Both p95 and p99 breach → 2 alerts
	if enq.count() != 2 {
		t.Errorf("expected 2 alerts, got %d", enq.count())
	}
}

func TestAlertManager_NoBreach(t *testing.T) {
	enq := &mockEnqueuer{}
	am := NewAlertManager(AlertConfig{
		P95Threshold: 200 * time.Millisecond,
		P99Threshold: 500 * time.Millisecond,
		CooldownSec:  1,
	}, enq, &mockLogger{})

	stats := Stats{
		P95:   100 * time.Millisecond,
		P99:   300 * time.Millisecond,
		Count: 100,
	}

	am.Check(stats)

	if enq.count() != 0 {
		t.Errorf("expected 0 alerts, got %d", enq.count())
	}
}

func TestAlertManager_Cooldown(t *testing.T) {
	enq := &mockEnqueuer{}
	am := NewAlertManager(AlertConfig{
		P95Threshold: 200 * time.Millisecond,
		P99Threshold: 500 * time.Millisecond,
		CooldownSec:  300, // 5 minutes
	}, enq, &mockLogger{})

	stats := Stats{
		P95:   250 * time.Millisecond,
		P99:   400 * time.Millisecond,
		Count: 100,
	}

	am.Check(stats)
	am.Check(stats) // second call — should be debounced

	if enq.count() != 1 {
		t.Errorf("expected 1 alert (cooldown), got %d", enq.count())
	}
}

func TestAlertManager_NilEnqueuer(t *testing.T) {
	am := NewAlertManager(AlertConfig{
		P95Threshold: 200 * time.Millisecond,
		P99Threshold: 500 * time.Millisecond,
		CooldownSec:  1,
	}, nil, &mockLogger{})

	stats := Stats{
		P95:   250 * time.Millisecond,
		P99:   600 * time.Millisecond,
		Count: 100,
	}

	// Should not panic
	am.Check(stats)
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/shared/infrastructure/metrics/latency/ -run TestAlertManager -v
```

Expected: FAIL — `NewAlertManager` not defined.

- [ ] **Step 3: Implement alert manager**

Create `internal/shared/infrastructure/metrics/latency/alert.go`:

```go
package latency

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gct/internal/shared/infrastructure/logger"
)

// TaskEnqueuer abstracts Asynq client for testability.
type TaskEnqueuer interface {
	EnqueueTask(ctx context.Context, taskType string, payload any, opts ...any) (any, error)
}

// AlertConfig configures latency alert thresholds.
type AlertConfig struct {
	P95Threshold time.Duration
	P99Threshold time.Duration
	CooldownSec  int
}

// AlertPayload is sent to the Telegram task queue.
type AlertPayload struct {
	MessageType string `json:"message_type"`
	Text        string `json:"text"`
}

// AlertManager checks latency stats against thresholds and sends Telegram alerts.
type AlertManager struct {
	cfg      AlertConfig
	enqueuer TaskEnqueuer
	log      logger.Log
	mu       sync.Mutex
	lastSent map[string]time.Time
}

// NewAlertManager creates a new alert manager.
func NewAlertManager(cfg AlertConfig, enqueuer TaskEnqueuer, l logger.Log) *AlertManager {
	if cfg.CooldownSec <= 0 {
		cfg.CooldownSec = 300
	}
	return &AlertManager{
		cfg:      cfg,
		enqueuer: enqueuer,
		log:      l,
		lastSent: make(map[string]time.Time),
	}
}

// Check evaluates stats against thresholds and sends alerts if breached.
func (a *AlertManager) Check(stats Stats) {
	if stats.Count == 0 {
		return
	}

	if a.cfg.P95Threshold > 0 && stats.P95 > a.cfg.P95Threshold {
		a.log.Warnw("p95 latency threshold breached",
			"p95", stats.P95.String(),
			"threshold", a.cfg.P95Threshold.String(),
			"count", stats.Count,
		)
		a.sendAlert("p95", stats.P95, a.cfg.P95Threshold, stats)
	}

	if a.cfg.P99Threshold > 0 && stats.P99 > a.cfg.P99Threshold {
		a.log.Errorw("p99 latency threshold breached",
			"p99", stats.P99.String(),
			"threshold", a.cfg.P99Threshold.String(),
			"count", stats.Count,
		)
		a.sendAlert("p99", stats.P99, a.cfg.P99Threshold, stats)
	}
}

func (a *AlertManager) sendAlert(percentile string, value, threshold time.Duration, stats Stats) {
	if a.enqueuer == nil {
		return
	}

	if a.isCooledDown(percentile) {
		return
	}

	text := fmt.Sprintf("⚠️ Latency Alert [%s]\n%s: %s (threshold: %s)\np50: %s | p95: %s | p99: %s\nRequests: %d | Window: %s",
		time.Now().UTC().Format("15:04:05"),
		percentile, value, threshold,
		stats.P50, stats.P95, stats.P99,
		stats.Count, stats.Window,
	)

	payload := AlertPayload{
		MessageType: "latency_alert",
		Text:        text,
	}

	_, _ = a.enqueuer.EnqueueTask(context.Background(), "task:send_telegram", payload)
}

func (a *AlertManager) isCooledDown(key string) bool {
	cooldown := time.Duration(a.cfg.CooldownSec) * time.Second

	a.mu.Lock()
	defer a.mu.Unlock()

	if last, ok := a.lastSent[key]; ok {
		if time.Since(last) < cooldown {
			return true
		}
	}
	a.lastSent[key] = time.Now()
	return false
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/shared/infrastructure/metrics/latency/ -run TestAlertManager -v
```

Expected: All 5 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/shared/infrastructure/metrics/latency/alert.go internal/shared/infrastructure/metrics/latency/alert_test.go
git commit -m "feat(latency): implement alert manager with Telegram + cooldown"
```

---

### Task 5: Implement Reporter

**Files:**
- Create: `internal/shared/infrastructure/metrics/latency/reporter.go`
- Create: `internal/shared/infrastructure/metrics/latency/reporter_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/shared/infrastructure/metrics/latency/reporter_test.go`:

```go
package latency

import (
	"context"
	"testing"
	"time"
)

func TestReporter_StartsAndStops(t *testing.T) {
	tr := NewTracker(300)
	enq := &mockEnqueuer{}
	am := NewAlertManager(AlertConfig{
		P95Threshold: 200 * time.Millisecond,
		P99Threshold: 500 * time.Millisecond,
		CooldownSec:  1,
	}, enq, &mockLogger{})

	r := NewReporter(tr, am, 50*time.Millisecond, &mockLogger{})

	ctx, cancel := context.WithCancel(context.Background())
	r.Start(ctx)

	// Record some data
	tr.Record(300 * time.Millisecond)

	// Wait for at least one tick
	time.Sleep(100 * time.Millisecond)

	cancel()
	r.Stop()

	// Should have triggered p95 alert
	if enq.count() == 0 {
		t.Error("expected at least 1 alert from reporter")
	}
}

func TestReporter_ResetsAfterWindow(t *testing.T) {
	tr := NewTracker(1) // 1-second window
	enq := &mockEnqueuer{}
	am := NewAlertManager(AlertConfig{
		P95Threshold: 200 * time.Millisecond,
		P99Threshold: 500 * time.Millisecond,
		CooldownSec:  1,
	}, enq, &mockLogger{})

	r := NewReporter(tr, am, 50*time.Millisecond, &mockLogger{})

	ctx, cancel := context.WithCancel(context.Background())
	r.Start(ctx)

	tr.Record(300 * time.Millisecond)

	// Wait for window to expire + tick
	time.Sleep(1200 * time.Millisecond)

	stats := tr.Stats()
	// After window reset, count should be 0
	if stats.Count != 0 {
		t.Errorf("Count after window reset = %d, want 0", stats.Count)
	}

	cancel()
	r.Stop()
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/shared/infrastructure/metrics/latency/ -run TestReporter -v
```

Expected: FAIL — `NewReporter` not defined.

- [ ] **Step 3: Implement reporter**

Create `internal/shared/infrastructure/metrics/latency/reporter.go`:

```go
package latency

import (
	"context"
	"sync"
	"time"

	"gct/internal/shared/infrastructure/logger"
)

// Reporter periodically logs latency stats and checks alert thresholds.
type Reporter struct {
	tracker  *Tracker
	alert    *AlertManager
	interval time.Duration
	log      logger.Log
	stopOnce sync.Once
	done     chan struct{}
}

// NewReporter creates a new periodic reporter.
func NewReporter(tracker *Tracker, alert *AlertManager, interval time.Duration, l logger.Log) *Reporter {
	return &Reporter{
		tracker:  tracker,
		alert:    alert,
		interval: interval,
		log:      l,
		done:     make(chan struct{}),
	}
}

// Start begins periodic reporting in a background goroutine.
func (r *Reporter) Start(ctx context.Context) {
	go r.run(ctx)
}

func (r *Reporter) run(ctx context.Context) {
	defer close(r.done)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	windowTicker := time.NewTicker(r.tracker.window)
	defer windowTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-windowTicker.C:
			r.tracker.Reset()
			r.log.Infow("latency tracker window reset", "window", r.tracker.window.String())
		case <-ticker.C:
			r.report()
		}
	}
}

func (r *Reporter) report() {
	stats := r.tracker.Stats()
	if stats.Count == 0 {
		return
	}

	r.log.Infow("latency stats",
		"p50", stats.P50.String(),
		"p95", stats.P95.String(),
		"p99", stats.P99.String(),
		"p999", stats.P999.String(),
		"mean", stats.Mean.String(),
		"max", stats.Max.String(),
		"count", stats.Count,
		"window", stats.Window,
	)

	r.alert.Check(stats)
}

// Stop waits for the reporter goroutine to finish.
func (r *Reporter) Stop() {
	r.stopOnce.Do(func() {
		<-r.done
	})
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/shared/infrastructure/metrics/latency/ -run TestReporter -v -timeout 10s
```

Expected: All 2 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/shared/infrastructure/metrics/latency/reporter.go internal/shared/infrastructure/metrics/latency/reporter_test.go
git commit -m "feat(latency): implement periodic reporter with window reset"
```

---

### Task 6: Implement Gin Middleware

**Files:**
- Create: `internal/shared/infrastructure/middleware/latency_tracker.go`
- Create: `internal/shared/infrastructure/middleware/latency_tracker_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/shared/infrastructure/middleware/latency_tracker_test.go`:

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/shared/infrastructure/metrics/latency"

	"github.com/gin-gonic/gin"
)

func TestLatencyTrackerMiddleware_RecordsLatency(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tracker := latency.NewTracker(300)

	router := gin.New()
	router.Use(LatencyTracker(tracker))
	router.GET("/test", func(c *gin.Context) {
		time.Sleep(5 * time.Millisecond)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	stats := tracker.Stats()
	if stats.Count != 1 {
		t.Errorf("Count = %d, want 1", stats.Count)
	}
	if stats.P50 < 5*time.Millisecond {
		t.Errorf("P50 = %v, expected >= 5ms", stats.P50)
	}
}

func TestLatencyTrackerMiddleware_MultipleRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tracker := latency.NewTracker(300)

	router := gin.New()
	router.Use(LatencyTracker(tracker))
	router.GET("/fast", func(c *gin.Context) { c.Status(http.StatusOK) })

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/fast", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	stats := tracker.Stats()
	if stats.Count != 10 {
		t.Errorf("Count = %d, want 10", stats.Count)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/shared/infrastructure/middleware/ -run TestLatencyTracker -v
```

Expected: FAIL — `LatencyTracker` not defined.

- [ ] **Step 3: Implement middleware**

Create `internal/shared/infrastructure/middleware/latency_tracker.go`:

```go
package middleware

import (
	"time"

	"gct/internal/shared/infrastructure/metrics/latency"

	"github.com/gin-gonic/gin"
)

// LatencyTracker returns a Gin middleware that records request latency to the HDR tracker.
func LatencyTracker(tracker *latency.Tracker) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		tracker.Record(time.Since(start))
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/shared/infrastructure/middleware/ -run TestLatencyTracker -v
```

Expected: All 2 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/shared/infrastructure/middleware/latency_tracker.go internal/shared/infrastructure/middleware/latency_tracker_test.go
git commit -m "feat(middleware): add latency tracker middleware"
```

---

### Task 7: Wire Into Application

**Files:**
- Modify: `internal/shared/infrastructure/middleware/setup.go:53-56`
- Modify: `internal/app/app.go:69-78`
- Modify: `internal/app/infra_routes.go:23-27`

- [ ] **Step 1: Add latency tracker to middleware setup**

In `internal/shared/infrastructure/middleware/setup.go`, add import and modify the metrics section.

Add to imports:

```go
"gct/internal/shared/infrastructure/metrics/latency"
```

Add `latencyTracker *latency.Tracker` parameter to `Setup` function signature:

```go
func Setup(handler *gin.Engine, cfg *config.Config, redisClient *redis.Client, bcMW *BCMiddleware, latencyTracker *latency.Tracker, l logger.Log) {
```

After the OTel metrics middleware block (after line 56), add:

```go
	// 3.2 In-memory latency tracking (HDR Histogram)
	if cfg.Metrics.LatencyEnabled && latencyTracker != nil {
		handler.Use(LatencyTracker(latencyTracker))
	}
```

- [ ] **Step 2: Initialize tracker, alert, reporter in app.go**

In `internal/app/app.go`, add import:

```go
"gct/internal/shared/infrastructure/metrics/latency"
```

After the metrics provider initialization block (after line 78), add:

```go
	// 1.1 Latency Percentile Tracker
	var latencyTracker *latency.Tracker
	var latencyReporter *latency.Reporter
	if cfg.Metrics.Enabled && cfg.Metrics.LatencyEnabled {
		latencyTracker = latency.NewTracker(cfg.Metrics.LatencyWindowSec)

		p95Threshold, _ := time.ParseDuration(cfg.Metrics.LatencyP95Threshold)
		p99Threshold, _ := time.ParseDuration(cfg.Metrics.LatencyP99Threshold)

		var alertEnqueuer latency.TaskEnqueuer
		if asynqClient != nil {
			alertEnqueuer = &asynqClientAdapter{asynqClient}
		}

		alertMgr := latency.NewAlertManager(latency.AlertConfig{
			P95Threshold: p95Threshold,
			P99Threshold: p99Threshold,
			CooldownSec:  300,
		}, alertEnqueuer, l)

		interval := time.Duration(cfg.Metrics.LatencyLogIntervalSec) * time.Second
		latencyReporter = latency.NewReporter(latencyTracker, alertMgr, interval, l)
		latencyReporter.Start(ctx)

		l.Infoc(ctx, "Latency percentile tracker enabled",
			"p95_threshold", cfg.Metrics.LatencyP95Threshold,
			"p99_threshold", cfg.Metrics.LatencyP99Threshold,
			"window_sec", cfg.Metrics.LatencyWindowSec,
			"log_interval_sec", cfg.Metrics.LatencyLogIntervalSec,
		)
	}
	_ = latencyReporter // used in shutdown below
```

Note: this block must be placed **before** the asynq client initialization (line 137), because `asynqClient` is declared earlier at line 140. Actually, `asynqClient` is declared at line 140 which is **after** line 78. So this latency block needs to be placed after both metrics provider AND asynq client are initialized. Move it to after line 145 (after `_ = asynqClient`):

```go
	// After line 145: _ = asynqClient

	// 3.1 Latency Percentile Tracker
	var latencyTracker *latency.Tracker
	var latencyReporter *latency.Reporter
	if cfg.Metrics.Enabled && cfg.Metrics.LatencyEnabled {
		latencyTracker = latency.NewTracker(cfg.Metrics.LatencyWindowSec)

		p95Threshold, _ := time.ParseDuration(cfg.Metrics.LatencyP95Threshold)
		p99Threshold, _ := time.ParseDuration(cfg.Metrics.LatencyP99Threshold)

		var alertEnqueuer latency.TaskEnqueuer
		if asynqClient != nil {
			alertEnqueuer = &asynqClientAdapter{asynqClient}
		}

		alertMgr := latency.NewAlertManager(latency.AlertConfig{
			P95Threshold: p95Threshold,
			P99Threshold: p99Threshold,
			CooldownSec:  300,
		}, alertEnqueuer, l)

		interval := time.Duration(cfg.Metrics.LatencyLogIntervalSec) * time.Second
		latencyReporter = latency.NewReporter(latencyTracker, alertMgr, interval, l)
		latencyReporter.Start(ctx)

		l.Infoc(ctx, "Latency percentile tracker enabled",
			"p95_threshold", cfg.Metrics.LatencyP95Threshold,
			"p99_threshold", cfg.Metrics.LatencyP99Threshold,
			"window_sec", cfg.Metrics.LatencyWindowSec,
			"log_interval_sec", cfg.Metrics.LatencyLogIntervalSec,
		)
	}
	_ = latencyReporter
```

- [ ] **Step 3: Pass latencyTracker to initRouter and Setup**

In `internal/app/app.go`, update the `initRouter` call (line 324):

```go
handler := initRouter(cfg, dddBCs, redisclient, pg, sseHub, metricsProvider, latencyTracker, l)
```

Update `initRouter` function signature (line 342):

```go
func initRouter(cfg *config.Config, bcs *DDDBoundedContexts, redisClient *redis.Client, pg *postgres.Postgres, sseHub *sse.Hub, metricsProvider *metrics.Provider, latencyTracker *latency.Tracker, l logger.Log) *gin.Engine {
```

Update `sharedmw.Setup` call (line 370):

```go
sharedmw.Setup(handler, cfg, redisClient, bcMW, latencyTracker, l)
```

- [ ] **Step 4: Add /metrics/latency endpoint to infra_routes.go**

In `internal/app/infra_routes.go`, add import:

```go
"gct/internal/shared/infrastructure/metrics/latency"
```

Update `setupInfraRoutes` function signature to accept tracker:

```go
func setupInfraRoutes(handler *gin.Engine, cfg *config.Config, pool *pgxpool.Pool, redisClient *redis.Client, metricsHandler http.Handler, latencyTracker *latency.Tracker, minioClient *miniogo.Client) {
```

After the Prometheus metrics block (after line 27), add:

```go
	// Latency percentile stats (in-memory HDR Histogram)
	if cfg.Metrics.LatencyEnabled && latencyTracker != nil {
		handler.GET("/metrics/latency", func(c *gin.Context) {
			stats := latencyTracker.Stats()
			c.JSON(http.StatusOK, gin.H{
				"p50":    stats.P50.String(),
				"p95":    stats.P95.String(),
				"p99":    stats.P99.String(),
				"p999":   stats.P999.String(),
				"mean":   stats.Mean.String(),
				"max":    stats.Max.String(),
				"count":  stats.Count,
				"window": stats.Window,
			})
		})
	}
```

Update the `setupInfraRoutes` call in `initRouter` (line 377 area):

```go
setupInfraRoutes(handler, cfg, pg.Pool, redisClient, metricsHandler, latencyTracker, nil)
```

- [ ] **Step 5: Verify compilation**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 6: Commit**

```bash
git add internal/shared/infrastructure/middleware/setup.go internal/app/app.go internal/app/infra_routes.go
git commit -m "feat(latency): wire tracker, reporter, and /metrics/latency endpoint"
```

---

### Task 8: Add Grafana Dashboard + PromQL

**Files:**
- Create: `docs/grafana/latency-dashboard.json`

- [ ] **Step 1: Create Grafana dashboard JSON**

Create `docs/grafana/latency-dashboard.json`:

```json
{
  "dashboard": {
    "title": "HTTP Latency Percentiles",
    "uid": "latency-p95-p99",
    "tags": ["latency", "http", "slo"],
    "timezone": "browser",
    "refresh": "30s",
    "time": { "from": "now-1h", "to": "now" },
    "panels": [
      {
        "id": 1,
        "title": "HTTP Request Latency Percentiles",
        "type": "timeseries",
        "gridPos": { "h": 8, "w": 24, "x": 0, "y": 0 },
        "targets": [
          {
            "expr": "histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "p50"
          },
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "p95"
          },
          {
            "expr": "histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "p99"
          },
          {
            "expr": "histogram_quantile(0.999, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "p999"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "s",
            "thresholds": {
              "steps": [
                { "color": "green", "value": null },
                { "color": "yellow", "value": 0.2 },
                { "color": "red", "value": 0.5 }
              ]
            }
          }
        }
      },
      {
        "id": 2,
        "title": "p99 Latency by Endpoint",
        "type": "timeseries",
        "gridPos": { "h": 8, "w": 24, "x": 0, "y": 8 },
        "targets": [
          {
            "expr": "histogram_quantile(0.99, sum by (le, path) (rate(http_request_duration_seconds_bucket[5m])))",
            "legendFormat": "{{ path }}"
          }
        ],
        "fieldConfig": {
          "defaults": { "unit": "s" }
        }
      },
      {
        "id": 3,
        "title": "Request Rate",
        "type": "timeseries",
        "gridPos": { "h": 8, "w": 12, "x": 0, "y": 16 },
        "targets": [
          {
            "expr": "sum(rate(http_requests_total[5m]))",
            "legendFormat": "req/s"
          }
        ],
        "fieldConfig": {
          "defaults": { "unit": "reqps" }
        }
      },
      {
        "id": 4,
        "title": "Requests In-Flight",
        "type": "gauge",
        "gridPos": { "h": 8, "w": 12, "x": 12, "y": 16 },
        "targets": [
          {
            "expr": "sum(http_requests_in_flight)",
            "legendFormat": "in-flight"
          }
        ]
      }
    ],
    "annotations": {
      "list": [
        {
          "name": "p99 Threshold Breach",
          "datasource": "Prometheus",
          "expr": "histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m])) > 0.5",
          "tagKeys": "latency,alert",
          "titleFormat": "p99 > 500ms"
        }
      ]
    }
  }
}
```

- [ ] **Step 2: Commit**

```bash
mkdir -p docs/grafana
git add docs/grafana/latency-dashboard.json
git commit -m "docs(grafana): add latency percentile dashboard with PromQL"
```

---

### Task 9: Run All Tests + Verify

- [ ] **Step 1: Run all latency package tests**

```bash
go test ./internal/shared/infrastructure/metrics/latency/ -v -race
```

Expected: All tests PASS, no race conditions.

- [ ] **Step 2: Run middleware tests**

```bash
go test ./internal/shared/infrastructure/middleware/ -run TestLatencyTracker -v -race
```

Expected: All tests PASS.

- [ ] **Step 3: Run full build**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 4: Run full test suite**

```bash
go test ./... -count=1 2>&1 | tail -30
```

Expected: no regressions.
