package errorx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"gct/internal/kernel/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

// TaskEnqueuer abstracts Asynq client for testability.
type TaskEnqueuer interface {
	EnqueueTask(ctx context.Context, taskType string, payload any, opts ...any) (any, error)
}

// AlerterConfig configures the error alerter.
type AlerterConfig struct {
	MinSeverity    ErrorSeverity
	DebouncePeriod time.Duration
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
	lastSent map[string]time.Time
	rdb      redis.Cmdable
	log      logger.Log
}

// NewAlerter creates a new error alerter.
func NewAlerter(enqueuer TaskEnqueuer, cfg AlerterConfig, rdb redis.Cmdable, log logger.Log) *Alerter {
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
		rdb:      rdb,
		log:      log,
	}
}

// SendError implements the Reporter interface from logging.go.
func (a *Alerter) SendError(err error) error {
	var appErr *AppError
	if !errors.As(err, &appErr) {
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

	// SendError satisfies the Reporter interface (no ctx parameter), so both
	// the enqueue call and the Redis fallback must originate a fresh context.
	_, enqErr := a.enqueuer.EnqueueTask(context.Background(), "task:send_telegram", payload)
	if enqErr != nil {
		if a.rdb != nil {
			data, _ := json.Marshal(payload)
			if pushErr := a.rdb.LPush(context.Background(), "alerter:pending", data).Err(); pushErr != nil {
				if a.log != nil {
					a.log.Warnw("Alerter: Redis LPUSH failed, alert dropped", "error", pushErr)
				}
			}
		} else if a.log != nil {
			a.log.Warnw("Alerter: enqueue failed and Redis unavailable, alert dropped", "error", enqErr)
		}
	}
	return nil
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

// Cleanup removes expired debounce entries to prevent memory leaks.
func (a *Alerter) Cleanup() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for code, last := range a.lastSent {
		if time.Since(last) > a.cfg.DebouncePeriod*2 {
			delete(a.lastSent, code)
		}
	}
}

// StartCleanup runs periodic cleanup in background.
func (a *Alerter) StartCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(a.cfg.DebouncePeriod * 2)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.Cleanup()
			}
		}
	}()
}

// StartPendingLoop periodically re-enqueues pending alerts from Redis.
func (a *Alerter) StartPendingLoop(ctx context.Context) {
	if a.rdb == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.flushPending(ctx)
			}
		}
	}()
}

func (a *Alerter) flushPending(ctx context.Context) {
	entries, err := a.rdb.LRange(ctx, "alerter:pending", 0, 49).Result()
	if err != nil || len(entries) == 0 {
		return
	}

	a.rdb.LTrim(ctx, "alerter:pending", int64(len(entries)), -1)

	for _, raw := range entries {
		var payload AlertPayload
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			continue
		}
		_, enqErr := a.enqueuer.EnqueueTask(ctx, "task:send_telegram", payload)
		if enqErr != nil {
			// Re-queue on failure
			a.rdb.RPush(ctx, "alerter:pending", raw)
			if a.log != nil {
				a.log.Warnw("Alerter: re-enqueue failed, kept in pending", "error", enqErr)
			}
			return // stop on first failure
		}
	}
}
