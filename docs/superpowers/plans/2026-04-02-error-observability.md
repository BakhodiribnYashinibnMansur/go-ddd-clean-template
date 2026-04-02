# Error Observability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Error correlation (request_id), alerting (Telegram), circuit breaker, error rate monitoring, health check endpoint'lar qo'shish.

**Architecture:** Mavjud request_id middleware'ni kuchaytirish, Reporter interface'ni Telegram+Asynq bilan implement qilish, circuit breaker pattern tashqi service'lar uchun, sliding window rate monitor, va /health + /ready endpoint'lar.

**Tech Stack:** Gin (mavjud), Zap logger (mavjud), Asynq (mavjud), Telegram client (mavjud), Redis (mavjud)

**Mavjud infra:**
- Request ID middleware allaqachon `middleware/logger.go` da — UUID yaratadi, `X-Request-ID` header'da qaytaradi, context'ga qo'yadi
- `contextx.GetRequestID(ctx)` — context'dan request_id olish
- Logger context-aware — `l.Errorc(ctx, ...)` kontextdan field'larni avtomatik oladi
- `Reporter` interface — `errors/logging.go` da defined, lekin implement qilinmagan
- `ErrorHookManager` — hook'lar qo'shish mumkin
- Asynq task queue — Telegram task handler allaqachon tayyor (oldingi feature'dan)

---

## File Structure

```
internal/shared/infrastructure/
├── errors/
│   ├── alerter.go              ← YANGI: Telegram alert Reporter implementation
│   ├── alerter_test.go         ← YANGI: alerter tests
│   ├── rate_monitor.go         ← YANGI: sliding window error rate monitor
│   ├── rate_monitor_test.go    ← YANGI: rate monitor tests
│   └── logging.go              (mavjud — SetReporter chaqirish)
├── circuitbreaker/
│   ├── breaker.go              ← YANGI: circuit breaker implementation
│   └── breaker_test.go         ← YANGI: circuit breaker tests
├── middleware/
│   └── logger.go               (mavjud — request_id allaqachon bor, OTel span attribute qo'shiladi)
├── httpx/response/
│   └── error.go                (mavjud — requestId ni context'dan olish)
internal/app/
├── init_health.go              ← YANGI: health check handler
├── app.go                      (mavjud — health route + alerter init)
```

---

### Task 1: Error Response'da request_id ni context'dan olish

Hozir `parseErrorToResponse()` `X-Request-ID` header'dan oladi yoki yangi UUID yaratadi. Lekin middleware allaqachon context'ga qo'ygan. Context'dan olishga o'tkazish kerak.

**Files:**
- Modify: `internal/shared/infrastructure/httpx/response/error.go`

- [ ] **Step 1: Update requestId resolution in parseErrorToResponse**

`internal/shared/infrastructure/httpx/response/error.go` da `parseErrorToResponse` funksiyasida request ID olish qismini o'zgartirish:

```go
// Import qo'shish:
"gct/internal/shared/infrastructure/contextx"

// Mavjud kodni o'zgartirish — request_id qismini:
// BEFORE:
// reqID := c.GetHeader(consts.HeaderXRequestID)
// if reqID == "" {
//     reqID = uuid.New().String()
// }

// AFTER:
reqID := contextx.GetRequestID(c.Request.Context())
if reqID == "" {
    reqID = c.GetHeader(consts.HeaderXRequestID)
}
if reqID == "" {
    reqID = uuid.New().String()
}
```

- [ ] **Step 2: Add trace_id from OTel span to response (optional enrichment)**

`internal/shared/infrastructure/middleware/logger.go` da OTel span'ga request_id attribute qo'shish:

```go
// Import:
"go.opentelemetry.io/otel/attribute"
"go.opentelemetry.io/otel/trace"

// Logger middleware ichida, request_id yaratilgandan keyin:
if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
    span.SetAttributes(attribute.String("request_id", requestID))
}
```

- [ ] **Step 3: Verify build and existing tests pass**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./... && go test ./internal/shared/infrastructure/httpx/response/... -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/shared/infrastructure/httpx/response/error.go internal/shared/infrastructure/middleware/logger.go
git commit -m "feat: resolve request_id from context in error response, add OTel span attribute"
```

---

### Task 2: Circuit Breaker

**Files:**
- Create: `internal/shared/infrastructure/circuitbreaker/breaker.go`
- Create: `internal/shared/infrastructure/circuitbreaker/breaker_test.go`

- [ ] **Step 1: Write circuit breaker tests**

```go
// internal/shared/infrastructure/circuitbreaker/breaker_test.go
package circuitbreaker_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/shared/infrastructure/circuitbreaker"
)

func TestBreaker_ClosedState_AllowsRequests(t *testing.T) {
	cb := circuitbreaker.New("test", circuitbreaker.Config{
		FailureThreshold: 3,
		Timeout:          1 * time.Second,
	})

	err := cb.Execute(func() error { return nil })
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cb.State() != circuitbreaker.StateClosed {
		t.Fatalf("expected CLOSED, got %s", cb.State())
	}
}

func TestBreaker_OpensAfterThreshold(t *testing.T) {
	cb := circuitbreaker.New("test", circuitbreaker.Config{
		FailureThreshold: 3,
		Timeout:          1 * time.Second,
	})

	fail := errors.New("service down")
	for i := 0; i < 3; i++ {
		_ = cb.Execute(func() error { return fail })
	}

	if cb.State() != circuitbreaker.StateOpen {
		t.Fatalf("expected OPEN after 3 failures, got %s", cb.State())
	}

	err := cb.Execute(func() error { return nil })
	if !errors.Is(err, circuitbreaker.ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cb := circuitbreaker.New("test", circuitbreaker.Config{
		FailureThreshold: 2,
		Timeout:          50 * time.Millisecond,
	})

	fail := errors.New("service down")
	_ = cb.Execute(func() error { return fail })
	_ = cb.Execute(func() error { return fail })

	if cb.State() != circuitbreaker.StateOpen {
		t.Fatalf("expected OPEN, got %s", cb.State())
	}

	time.Sleep(60 * time.Millisecond)

	// Should transition to HALF_OPEN and allow one request
	err := cb.Execute(func() error { return nil })
	if err != nil {
		t.Fatalf("expected success in HALF_OPEN, got %v", err)
	}
	if cb.State() != circuitbreaker.StateClosed {
		t.Fatalf("expected CLOSED after successful HALF_OPEN request, got %s", cb.State())
	}
}

func TestBreaker_HalfOpenFailureReOpens(t *testing.T) {
	cb := circuitbreaker.New("test", circuitbreaker.Config{
		FailureThreshold: 1,
		Timeout:          50 * time.Millisecond,
	})

	fail := errors.New("down")
	_ = cb.Execute(func() error { return fail })
	time.Sleep(60 * time.Millisecond)

	// HALF_OPEN: one request fails → back to OPEN
	err := cb.Execute(func() error { return fail })
	if err == nil {
		t.Fatal("expected error in HALF_OPEN failure")
	}
	if cb.State() != circuitbreaker.StateOpen {
		t.Fatalf("expected OPEN after HALF_OPEN failure, got %s", cb.State())
	}
}

func TestBreaker_SuccessResetsFailureCount(t *testing.T) {
	cb := circuitbreaker.New("test", circuitbreaker.Config{
		FailureThreshold: 3,
		Timeout:          1 * time.Second,
	})

	fail := errors.New("fail")
	_ = cb.Execute(func() error { return fail })
	_ = cb.Execute(func() error { return fail })
	// 2 failures, then success should reset
	_ = cb.Execute(func() error { return nil })

	if cb.State() != circuitbreaker.StateClosed {
		t.Fatalf("expected CLOSED after success reset, got %s", cb.State())
	}

	// Should need 3 more failures to open
	_ = cb.Execute(func() error { return fail })
	_ = cb.Execute(func() error { return fail })
	if cb.State() != circuitbreaker.StateClosed {
		t.Fatalf("expected still CLOSED after 2 failures, got %s", cb.State())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/shared/infrastructure/circuitbreaker/... -v`
Expected: FAIL — package not found

- [ ] **Step 3: Implement circuit breaker**

```go
// internal/shared/infrastructure/circuitbreaker/breaker.go
package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when the circuit is open and requests are blocked.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// State represents the circuit breaker state.
type State string

const (
	StateClosed   State = "CLOSED"
	StateOpen     State = "OPEN"
	StateHalfOpen State = "HALF_OPEN"
)

// Config holds circuit breaker configuration.
type Config struct {
	FailureThreshold int           // failures before opening (default: 5)
	Timeout          time.Duration // how long to stay open before half-open (default: 30s)
}

// Breaker implements the circuit breaker pattern.
type Breaker struct {
	name     string
	cfg      Config
	mu       sync.Mutex
	state    State
	failures int
	lastFail time.Time
}

// New creates a new circuit breaker.
func New(name string, cfg Config) *Breaker {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 5
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	return &Breaker{
		name:  name,
		cfg:   cfg,
		state: StateClosed,
	}
}

// State returns the current state.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.state == StateOpen && time.Since(b.lastFail) > b.cfg.Timeout {
		b.state = StateHalfOpen
	}
	return b.state
}

// Name returns the breaker name.
func (b *Breaker) Name() string {
	return b.name
}

// Execute runs fn through the circuit breaker.
func (b *Breaker) Execute(fn func() error) error {
	b.mu.Lock()

	// Check if open → half-open transition
	if b.state == StateOpen {
		if time.Since(b.lastFail) > b.cfg.Timeout {
			b.state = StateHalfOpen
		} else {
			b.mu.Unlock()
			return ErrCircuitOpen
		}
	}

	currentState := b.state
	b.mu.Unlock()

	err := fn()

	b.mu.Lock()
	defer b.mu.Unlock()

	if err != nil {
		b.failures++
		b.lastFail = time.Now()

		if currentState == StateHalfOpen {
			b.state = StateOpen
		} else if b.failures >= b.cfg.FailureThreshold {
			b.state = StateOpen
		}
		return err
	}

	// Success
	b.failures = 0
	b.state = StateClosed
	return nil
}
```

- [ ] **Step 4: Run tests**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/shared/infrastructure/circuitbreaker/... -v`
Expected: PASS (5 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/shared/infrastructure/circuitbreaker/
git commit -m "feat: add circuit breaker pattern implementation"
```

---

### Task 3: Error Alerter (Telegram via Asynq)

**Files:**
- Create: `internal/shared/infrastructure/errors/alerter.go`
- Create: `internal/shared/infrastructure/errors/alerter_test.go`
- Modify: `internal/app/init_asynq.go` — alerter init

- [ ] **Step 1: Write alerter tests**

```go
// internal/shared/infrastructure/errors/alerter_test.go
package errors_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	apperrors "gct/internal/shared/infrastructure/errors"
)

type mockEnqueuer struct {
	mu       sync.Mutex
	tasks    []string
	enqErr   error
}

func (m *mockEnqueuer) EnqueueTask(ctx context.Context, taskType string, payload any, opts ...any) (any, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, _ := json.Marshal(payload)
	m.tasks = append(m.tasks, taskType+":"+string(b))
	return nil, m.enqErr
}

func TestAlerter_SendsCriticalErrors(t *testing.T) {
	enq := &mockEnqueuer{}
	alerter := apperrors.NewAlerter(enq, apperrors.AlerterConfig{
		MinSeverity:    apperrors.SeverityCritical,
		DebouncePeriod: 0, // no debounce for test
	})

	err := apperrors.New(apperrors.ErrRepoConnection, "")
	alerter.SendError(err)

	enq.mu.Lock()
	defer enq.mu.Unlock()
	if len(enq.tasks) != 1 {
		t.Fatalf("expected 1 task enqueued, got %d", len(enq.tasks))
	}
}

func TestAlerter_SkipsLowSeverity(t *testing.T) {
	enq := &mockEnqueuer{}
	alerter := apperrors.NewAlerter(enq, apperrors.AlerterConfig{
		MinSeverity:    apperrors.SeverityHigh,
		DebouncePeriod: 0,
	})

	err := apperrors.New(apperrors.ErrBadRequest, "") // LOW severity
	alerter.SendError(err)

	enq.mu.Lock()
	defer enq.mu.Unlock()
	if len(enq.tasks) != 0 {
		t.Fatalf("expected 0 tasks for low severity, got %d", len(enq.tasks))
	}
}

func TestAlerter_DebouncesSameCode(t *testing.T) {
	enq := &mockEnqueuer{}
	alerter := apperrors.NewAlerter(enq, apperrors.AlerterConfig{
		MinSeverity:    apperrors.SeverityCritical,
		DebouncePeriod: 100 * time.Millisecond,
	})

	err := apperrors.New(apperrors.ErrRepoConnection, "")
	alerter.SendError(err)
	alerter.SendError(err) // should be debounced
	alerter.SendError(err) // should be debounced

	enq.mu.Lock()
	count := len(enq.tasks)
	enq.mu.Unlock()
	if count != 1 {
		t.Fatalf("expected 1 task (debounced), got %d", count)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/shared/infrastructure/errors/... -run TestAlerter -v`
Expected: FAIL

- [ ] **Step 3: Implement alerter**

```go
// internal/shared/infrastructure/errors/alerter.go
package errors

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TaskEnqueuer abstracts Asynq client for testability.
type TaskEnqueuer interface {
	EnqueueTask(ctx context.Context, taskType string, payload any, opts ...any) (any, error)
}

// AlerterConfig configures the error alerter.
type AlerterConfig struct {
	MinSeverity    ErrorSeverity // minimum severity to alert (default: HIGH)
	DebouncePeriod time.Duration // debounce same error code (default: 1 minute)
}

// AlertPayload is sent to the Telegram task queue.
type AlertPayload struct {
	MessageType string `json:"message_type"`
	Text        string `json:"text"`
}

// Alerter sends error alerts via Asynq task queue.
type Alerter struct {
	enqueuer TaskEnqueuer
	cfg      AlerterConfig
	mu       sync.Mutex
	lastSent map[string]time.Time // code → last alert time
}

// NewAlerter creates a new error alerter.
func NewAlerter(enqueuer TaskEnqueuer, cfg AlerterConfig) *Alerter {
	if cfg.MinSeverity == "" {
		cfg.MinSeverity = SeverityHigh
	}
	if cfg.DebouncePeriod == 0 {
		cfg.DebouncePeriod = time.Minute
	}
	return &Alerter{
		enqueuer: enqueuer,
		cfg:      cfg,
		lastSent: make(map[string]time.Time),
	}
}

// SendError implements the Reporter interface.
func (a *Alerter) SendError(err error) error {
	var appErr *AppError
	if !As(err, &appErr) {
		return nil
	}

	severity := GetSeverity(appErr.Type)
	if !a.shouldAlert(severity) {
		return nil
	}

	if a.isDebounced(appErr.Type) {
		return nil
	}

	text := fmt.Sprintf("🚨 %s [%s]\nCode: %s (%s)\nSeverity: %s\nCategory: %s\nMessage: %s",
		severity, time.Now().UTC().Format("15:04:05"),
		appErr.Type, appErr.Code,
		severity, GetCategory(appErr.Type),
		appErr.UserMsg,
	)
	if appErr.Details != "" {
		text += "\nDetails: " + appErr.Details
	}

	payload := AlertPayload{
		MessageType: "error",
		Text:        text,
	}

	_, _ = a.enqueuer.EnqueueTask(context.Background(), "task:send_telegram", payload)
	return nil
}

// As extracts AppError from error (helper to avoid import cycle).
func As(err error, target **AppError) bool {
	return asAppError(err, target)
}

func asAppError(err error, target **AppError) bool {
	if e, ok := err.(*AppError); ok {
		*target = e
		return true
	}
	return false
}

func (a *Alerter) shouldAlert(severity ErrorSeverity) bool {
	order := map[ErrorSeverity]int{
		SeverityInfo:     0,
		SeverityLow:      1,
		SeverityMedium:   2,
		SeverityHigh:     3,
		SeverityCritical: 4,
	}
	return order[severity] >= order[a.cfg.MinSeverity]
}

func (a *Alerter) isDebounced(code string) bool {
	if a.cfg.DebouncePeriod <= 0 {
		return false
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	if last, ok := a.lastSent[code]; ok {
		if time.Since(last) < a.cfg.DebouncePeriod {
			return true
		}
	}
	a.lastSent[code] = time.Now()
	return false
}
```

- [ ] **Step 4: Run tests**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/shared/infrastructure/errors/... -run TestAlerter -v`
Expected: PASS (3 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/shared/infrastructure/errors/alerter.go internal/shared/infrastructure/errors/alerter_test.go
git commit -m "feat: add error alerter with Telegram via Asynq, debounce support"
```

---

### Task 4: Error Rate Monitor

**Files:**
- Create: `internal/shared/infrastructure/errors/rate_monitor.go`
- Create: `internal/shared/infrastructure/errors/rate_monitor_test.go`

- [ ] **Step 1: Write rate monitor tests**

```go
// internal/shared/infrastructure/errors/rate_monitor_test.go
package errors_test

import (
	"testing"
	"time"

	apperrors "gct/internal/shared/infrastructure/errors"
)

func TestRateMonitor_DetectsThresholdBreach(t *testing.T) {
	var alertedCode string
	monitor := apperrors.NewRateMonitor(apperrors.RateMonitorConfig{
		Window:    1 * time.Second,
		Threshold: 3,
		OnBreach: func(code string, count int) {
			alertedCode = code
		},
	})

	for i := 0; i < 4; i++ {
		monitor.Record("TEST_ERROR")
	}

	if alertedCode != "TEST_ERROR" {
		t.Fatalf("expected breach alert for TEST_ERROR, got %q", alertedCode)
	}
}

func TestRateMonitor_NoBreach_UnderThreshold(t *testing.T) {
	breached := false
	monitor := apperrors.NewRateMonitor(apperrors.RateMonitorConfig{
		Window:    1 * time.Second,
		Threshold: 10,
		OnBreach: func(code string, count int) {
			breached = true
		},
	})

	for i := 0; i < 5; i++ {
		monitor.Record("TEST_ERROR")
	}

	if breached {
		t.Fatal("should not breach at 5/10 threshold")
	}
}

func TestRateMonitor_WindowExpiry(t *testing.T) {
	breachCount := 0
	monitor := apperrors.NewRateMonitor(apperrors.RateMonitorConfig{
		Window:    50 * time.Millisecond,
		Threshold: 3,
		OnBreach: func(code string, count int) {
			breachCount++
		},
	})

	monitor.Record("TEST")
	monitor.Record("TEST")
	monitor.Record("TEST")
	// Should NOT breach yet — debounce within window
	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	// New window — 3 more should breach again
	monitor.Record("TEST")
	monitor.Record("TEST")
	monitor.Record("TEST")
	monitor.Record("TEST")

	if breachCount < 1 {
		t.Fatalf("expected at least 1 breach, got %d", breachCount)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/shared/infrastructure/errors/... -run TestRateMonitor -v`
Expected: FAIL

- [ ] **Step 3: Implement rate monitor**

```go
// internal/shared/infrastructure/errors/rate_monitor.go
package errors

import (
	"sync"
	"time"
)

// RateMonitorConfig configures the error rate monitor.
type RateMonitorConfig struct {
	Window    time.Duration               // sliding window size (default: 1 minute)
	Threshold int                         // max errors per window per code (default: 10)
	OnBreach  func(code string, count int) // called when threshold is breached
}

// RateMonitor tracks error rates per code using a sliding window.
type RateMonitor struct {
	cfg     RateMonitorConfig
	mu      sync.Mutex
	windows map[string]*window
}

type window struct {
	count     int
	start     time.Time
	breached  bool // already alerted for this window
}

// NewRateMonitor creates a new rate monitor.
func NewRateMonitor(cfg RateMonitorConfig) *RateMonitor {
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	if cfg.Threshold <= 0 {
		cfg.Threshold = 10
	}
	return &RateMonitor{
		cfg:     cfg,
		windows: make(map[string]*window),
	}
}

// Record records an error occurrence and checks the threshold.
func (m *RateMonitor) Record(code string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	w, ok := m.windows[code]
	if !ok || time.Since(w.start) > m.cfg.Window {
		m.windows[code] = &window{count: 1, start: time.Now()}
		return
	}

	w.count++
	if w.count > m.cfg.Threshold && !w.breached {
		w.breached = true
		if m.cfg.OnBreach != nil {
			m.cfg.OnBreach(code, w.count)
		}
	}
}

// RateMonitorHook returns an ErrorHook that feeds errors into the rate monitor.
func RateMonitorHook(monitor *RateMonitor) func(error) {
	return func(err error) {
		code := GetCode(err)
		if code != "" {
			monitor.Record(code)
		}
	}
}
```

- [ ] **Step 4: Run tests**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/shared/infrastructure/errors/... -run TestRateMonitor -v`
Expected: PASS (3 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/shared/infrastructure/errors/rate_monitor.go internal/shared/infrastructure/errors/rate_monitor_test.go
git commit -m "feat: add sliding window error rate monitor with breach callback"
```

---

### Task 5: Health Check Endpoints

**Files:**
- Create: `internal/app/init_health.go`
- Modify: `internal/app/app.go` — health route registration

- [ ] **Step 1: Create health check handler**

```go
// internal/app/init_health.go
package app

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type healthDeps struct {
	pgPool *pgxpool.Pool
	redis  *redis.Client
}

var startTime = time.Now()

func registerHealthRoutes(r *gin.Engine, deps healthDeps) {
	r.GET("/health", handleHealth)
	r.GET("/ready", handleReady(deps))
}

func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"uptime": time.Since(startTime).String(),
	})
}

func handleReady(deps healthDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		checks := make(map[string]string)
		allOK := true

		// PostgreSQL
		if deps.pgPool != nil {
			if err := deps.pgPool.Ping(ctx); err != nil {
				checks["postgres"] = "unhealthy: " + err.Error()
				allOK = false
			} else {
				checks["postgres"] = "ok"
			}
		} else {
			checks["postgres"] = "not configured"
		}

		// Redis
		if deps.redis != nil {
			if err := deps.redis.Ping(ctx).Err(); err != nil {
				checks["redis"] = "unhealthy: " + err.Error()
				allOK = false
			} else {
				checks["redis"] = "ok"
			}
		} else {
			checks["redis"] = "not configured"
		}

		status := "ok"
		statusCode := http.StatusOK
		if !allOK {
			status = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, gin.H{
			"status": status,
			"checks": checks,
			"uptime": time.Since(startTime).String(),
		})
	}
}
```

- [ ] **Step 2: Register health routes in app.go**

Find the `initRouter` function in `internal/app/app.go` and add health route registration. Look for where infrastructure routes are registered (swagger, metrics) and add:

```go
// Register health check routes
registerHealthRoutes(handler, healthDeps{
    pgPool: pg.Pool,
    redis:  redisClient,
})
```

- [ ] **Step 3: Verify build**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./...`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add internal/app/init_health.go internal/app/app.go
git commit -m "feat: add /health and /ready endpoints with postgres and redis checks"
```

---

### Task 6: Wire Alerter + Rate Monitor in App Initialization

**Files:**
- Modify: `internal/app/app.go` or `internal/app/init_asynq.go`

- [ ] **Step 1: Initialize alerter and rate monitor**

In `internal/app/app.go`, after Asynq client initialization, add:

```go
// Initialize error alerter (sends CRITICAL/HIGH errors to Telegram via Asynq)
if asynqClient != nil {
    alerter := apperrors.NewAlerter(asynqClient, apperrors.AlerterConfig{
        MinSeverity:    apperrors.SeverityHigh,
        DebouncePeriod: time.Minute,
    })
    apperrors.SetReporter(alerter)

    // Initialize error rate monitor
    rateMonitor := apperrors.NewRateMonitor(apperrors.RateMonitorConfig{
        Window:    time.Minute,
        Threshold: 10,
        OnBreach: func(code string, count int) {
            alerter.SendError(
                apperrors.New(apperrors.ErrInternal, "").
                    WithDetails(fmt.Sprintf("Error rate breach: %s occurred %d times in 1 minute", code, count)),
            )
        },
    })

    hookMgr := apperrors.GetGlobalHookManager()
    hookMgr.AddHook(apperrors.RateMonitorHook(rateMonitor))
}
```

Add imports: `apperrors "gct/internal/shared/infrastructure/errors"`, `"time"`, `"fmt"`

- [ ] **Step 2: Verify build and all tests pass**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./... && go test $(go list ./... | grep -v "test/e2e/flows/user/client") 2>&1 | grep FAIL`
Expected: no FAIL output

- [ ] **Step 3: Commit**

```bash
git add internal/app/app.go
git commit -m "feat: wire error alerter and rate monitor in app initialization"
```

---

### Task 7: Full Build and Test Verification

- [ ] **Step 1: Build entire project**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./...`
Expected: no errors

- [ ] **Step 2: Run all tests**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test $(go list ./... | grep -v "test/e2e/flows/user/client") 2>&1 | grep FAIL`
Expected: no FAIL lines

- [ ] **Step 3: Run go vet**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go vet ./...`
Expected: no issues
