package errors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookConfig configures a webhook notification target.
type WebhookConfig struct {
	URL     string
	Headers map[string]string
	Timeout time.Duration
}

// WebhookReporter sends error alerts to a webhook URL (Slack, Discord, PagerDuty, etc).
type WebhookReporter struct {
	cfg    WebhookConfig
	client *http.Client
}

// NewWebhookReporter creates a new webhook reporter.
func NewWebhookReporter(cfg WebhookConfig) *WebhookReporter {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &WebhookReporter{
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.Timeout},
	}
}

// SendError sends error details to the configured webhook.
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

	ctx, cancel := context.WithTimeout(context.Background(), w.cfg.Timeout)
	defer cancel()

	req, err2 := http.NewRequestWithContext(ctx, http.MethodPost, w.cfg.URL, bytes.NewReader(body))
	if err2 != nil {
		return fmt.Errorf("create webhook request: %w", err2)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range w.cfg.Headers {
		req.Header.Set(k, v)
	}

	resp, err2 := w.client.Do(req)
	if err2 != nil {
		return fmt.Errorf("send webhook: %w", err2)
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
