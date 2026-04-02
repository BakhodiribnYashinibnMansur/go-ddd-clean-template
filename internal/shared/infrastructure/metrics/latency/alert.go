package latency

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TaskEnqueuer abstracts Asynq task enqueueing for testability.
type TaskEnqueuer interface {
	EnqueueTask(ctx context.Context, taskType string, payload any, opts ...any) (any, error)
}

// AlertConfig configures latency alert thresholds and cooldown.
type AlertConfig struct {
	P95Threshold time.Duration
	P99Threshold time.Duration
	Cooldown     time.Duration
}

// AlertPayload is the Telegram message payload.
type AlertPayload struct {
	MessageType string `json:"message_type"`
	Text        string `json:"text"`
}

// AlertManager checks latency stats against thresholds and sends alerts.
type AlertManager struct {
	enqueuer TaskEnqueuer
	cfg      AlertConfig
	mu       sync.Mutex
	lastSent time.Time
}

// NewAlertManager creates a new alert manager.
func NewAlertManager(enqueuer TaskEnqueuer, cfg AlertConfig) *AlertManager {
	if cfg.Cooldown == 0 {
		cfg.Cooldown = 5 * time.Minute
	}
	return &AlertManager{
		enqueuer: enqueuer,
		cfg:      cfg,
	}
}

// Check evaluates stats against thresholds and sends an alert if breached.
func (a *AlertManager) Check(stats LatencyStats) {
	if a == nil || a.enqueuer == nil {
		return
	}
	if stats.Count == 0 {
		return
	}

	var reasons []string
	if a.cfg.P95Threshold > 0 && stats.P95 > a.cfg.P95Threshold {
		reasons = append(reasons, fmt.Sprintf("p95=%v > %v", stats.P95, a.cfg.P95Threshold))
	}
	if a.cfg.P99Threshold > 0 && stats.P99 > a.cfg.P99Threshold {
		reasons = append(reasons, fmt.Sprintf("p99=%v > %v", stats.P99, a.cfg.P99Threshold))
	}

	if len(reasons) == 0 {
		return
	}
	if a.isCooledDown() {
		return
	}

	a.sendAlert(stats, reasons)
}

func (a *AlertManager) isCooledDown() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.lastSent.IsZero() && time.Since(a.lastSent) < a.cfg.Cooldown {
		return true
	}
	return false
}

func (a *AlertManager) sendAlert(stats LatencyStats, reasons []string) {
	a.mu.Lock()
	a.lastSent = time.Now()
	a.mu.Unlock()

	text := fmt.Sprintf("⚠️ Latency Alert [%s]\np50=%v p95=%v p99=%v mean=%v count=%d\nReasons: %v",
		time.Now().UTC().Format("15:04:05"),
		stats.P50, stats.P95, stats.P99, stats.Mean, stats.Count, reasons)

	payload := AlertPayload{MessageType: "latency_alert", Text: text}
	_, _ = a.enqueuer.EnqueueTask(context.Background(), "task:send_telegram", payload)
}
