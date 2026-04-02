# Latency Percentile Tracking (p95/p99)

**Date:** 2026-04-02
**Status:** Approved

## Problem

Prometheus histogram bor, lekin:
- Real-time p95/p99 qiymatlarini kodda bilish imkoni yo'q (Prometheus scrape kutish kerak)
- Latency threshold oshganda avtomatik alert yo'q
- Periodic latency statistikasi log ga yozilmaydi

## Solution

HDR Histogram yordamida in-memory percentile tracking + Prometheus PromQL + Grafana dashboard.

## Architecture

```
┌─────────────────────────────────────────────────┐
│                   Middleware                      │
│              (latency_tracker.go)                 │
│  request → HDR record → check threshold → alert  │
└──────────────┬──────────────────────┬────────────┘
               │                      │
       ┌───────▼────────┐    ┌────────▼─────────┐
       │  LatencyTracker │    │  AlertManager    │
       │  (tracker.go)   │    │  (alert.go)      │
       │                 │    │                   │
       │ • Record()      │    │ • Debounce        │
       │ • P95() / P99() │    │ • Asynq→Telegram  │
       │ • Stats()       │    │ • Cooldown period │
       │ • Reset()       │    └───────────────────┘
       └───────┬─────────┘
               │
       ┌───────▼────────┐
       │  Periodic Log   │
       │  (goroutine)    │
       │                 │
       │ har 60s:        │
       │ • p95, p99 log  │
       │ • threshold?    │
       │   → Telegram    │
       └────────────────┘

       ┌────────────────────┐
       │  /metrics/latency  │  ← REST endpoint
       │  JSON: p50,p95,    │
       │  p99,count,avg     │
       └────────────────────┘
```

## File Structure

```
internal/shared/infrastructure/metrics/latency/
├── tracker.go       # HDR Histogram wrapper — Record, P95, P99, Stats
├── alert.go         # Threshold check + Asynq → Telegram alert
└── reporter.go      # Periodic log + goroutine lifecycle

internal/shared/infrastructure/middleware/
├── latency_tracker.go   # Gin middleware — Record + threshold check
```

## Config

```go
type Metrics struct {
    // ... existing fields ...
    LatencyEnabled        bool   `yaml:"latency_enabled" env:"METRICS_LATENCY_ENABLED" envDefault:"true"`
    LatencyP95Threshold   string `yaml:"latency_p95_threshold" env:"METRICS_LATENCY_P95_THRESHOLD" envDefault:"200ms"`
    LatencyP99Threshold   string `yaml:"latency_p99_threshold" env:"METRICS_LATENCY_P99_THRESHOLD" envDefault:"500ms"`
    LatencyWindowSec      int    `yaml:"latency_window_sec" env:"METRICS_LATENCY_WINDOW_SEC" envDefault:"300"`
    LatencyLogIntervalSec int    `yaml:"latency_log_interval_sec" env:"METRICS_LATENCY_LOG_INTERVAL_SEC" envDefault:"60"`
}
```

- `LatencyP95Threshold` / `LatencyP99Threshold` — config dan o'qiladi, `time.Duration` formatida
- `LatencyWindowSec` — HDR histogram rotation oynasi (default 5 min)
- `LatencyLogIntervalSec` — har necha sekundda log ga yozish (default 60s)

## Components

### LatencyTracker (tracker.go)

```go
type Stats struct {
    P50     time.Duration `json:"p50"`
    P95     time.Duration `json:"p95"`
    P99     time.Duration `json:"p99"`
    P999    time.Duration `json:"p999"`
    Mean    time.Duration `json:"mean"`
    Max     time.Duration `json:"max"`
    Count   int64         `json:"count"`
    Window  string        `json:"window"`
}
```

- HDR Histogram: `1µs` dan `10s` gacha, 3 significant figures
- Thread-safe (`sync.Mutex`)
- `Reset()` har window tugaganda avtomatik chaqiriladi

### AlertManager (alert.go)

```go
type AlertConfig struct {
    P95Threshold  time.Duration
    P99Threshold  time.Duration
    CooldownSec   int  // default: 300s
}
```

- p95 > threshold → Warn log + Telegram
- p99 > threshold → Error log + Telegram
- Cooldown: bir xil alert 5 daqiqa ichida qayta yuborilmaydi
- Mavjud `TaskEnqueuer` interface orqali Asynq ga yuboradi

### Reporter (reporter.go)

- Har `interval` da: Stats oladi → log ga yozadi → AlertManager.Check()
- `ctx.Done()` da to'xtaydi

### Middleware (latency_tracker.go)

- `c.Next()` dan keyin `tracker.Record(latency)`
- Mavjud `OTelMetrics` bilan birga ishlaydi (biri Prometheus uchun, biri in-memory uchun)

### REST Endpoint

```
GET /metrics/latency
```

Response:
```json
{
    "p50": "12ms",
    "p95": "145ms",
    "p99": "380ms",
    "p999": "1.2s",
    "mean": "25ms",
    "max": "2.1s",
    "count": 15420,
    "window": "5m"
}
```

## Prometheus PromQL

```promql
# p95
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# p99
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))

# endpoint bo'yicha p99
histogram_quantile(0.99, sum by (le, http_path) (rate(http_request_duration_seconds_bucket[5m])))
```

## Grafana Dashboard

`docs/grafana/latency-dashboard.json` — tayyor import qilinadigan dashboard:
- p50/p95/p99 trend panellari
- Endpoint bo'yicha p99 breakdown
- Alert rules (p99 > threshold)

## Dependency

```
go get github.com/HdrHistogram/hdrhistogram-go
```

## Integration Points

1. **Middleware** — `setup.go` da `OTelMetrics` dan keyin ro'yxatga olinadi
2. **App init** — `app.go` da tracker, alert manager, reporter yaratiladi
3. **Infra routes** — `/metrics/latency` endpoint qo'shiladi
4. **Asynq** — mavjud Telegram task pipeline orqali alert yuboriladi
