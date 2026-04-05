package errorx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gct/internal/kernel/infrastructure/circuitbreaker"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

const redisWebhookPendingKey = "webhook:pending"

// WebhookConfig configures a webhook notification target.
type WebhookConfig struct {
	URL     string
	Headers map[string]string
	Timeout time.Duration
}

// WebhookReporter sends error alerts to a webhook URL (Slack, Discord, PagerDuty, etc).
type WebhookReporter struct {
	cfg       WebhookConfig
	client    *http.Client
	webhookCB *circuitbreaker.Breaker
	rdb       redis.Cmdable
	log       logger.Log
}

// NewWebhookReporter creates a new webhook reporter.
func NewWebhookReporter(cfg WebhookConfig, rdb redis.Cmdable, log logger.Log) *WebhookReporter {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &WebhookReporter{
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.Timeout},
		rdb:    rdb,
		log:    log,
		webhookCB: circuitbreaker.New("webhook", circuitbreaker.Config{
			FailureThreshold: 3,
			Timeout:          60 * time.Second,
		}),
	}
}

// SendError enqueues error details to Redis for async webhook delivery.
func (w *WebhookReporter) SendError(err error) error {
	var appErr *AppError
	if !asAppError(err, &appErr) {
		return nil
	}

	payload := map[string]any{
		"code":      appErr.Type,
		"numeric":   appErr.Code,
		"severity":  string(GetSeverity(appErr.Type)),
		"category":  string(GetCategory(appErr.Type)),
		"message":   appErr.UserMsg,
		"details":   appErr.Details,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	body, err2 := json.Marshal(payload)
	if err2 != nil {
		return fmt.Errorf("marshal webhook payload: %w", err2)
	}

	if w.rdb == nil {
		if w.log != nil {
			w.log.Warnw("Webhook: Redis unavailable, error dropped")
		}
		return nil
	}

	// SendError's signature does not accept a ctx (it satisfies the Reporter
	// interface from logging.go), so we must originate one here.
	if pushErr := w.rdb.LPush(context.Background(), redisWebhookPendingKey, body).Err(); pushErr != nil {
		if w.log != nil {
			w.log.Warnw("Webhook: Redis LPUSH failed, error dropped", "error", pushErr)
		}
		return nil
	}
	return nil
}

// doPost sends a raw JSON body to the configured webhook URL.
func (w *WebhookReporter) doPost(body []byte) error {
	// doPost is invoked from a background pending-queue drain that has no
	// caller context; the HTTP call must not be tied to any request lifetime.
	ctx, cancel := context.WithTimeout(context.Background(), w.cfg.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.cfg.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range w.cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// asAppError extracts AppError from error.
func asAppError(err error, target **AppError) bool {
	if e, ok := err.(*AppError); ok {
		*target = e
		return true
	}
	return false
}
