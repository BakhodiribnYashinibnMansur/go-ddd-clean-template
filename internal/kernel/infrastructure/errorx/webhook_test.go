package errorx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/kernel/infrastructure/logger"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

)

// ---------------------------------------------------------------------------
// WebhookReporter tests
// ---------------------------------------------------------------------------

func TestNewWebhookReporter(t *testing.T) {
	cfg := WebhookConfig{URL: "http://example.com/hook"}
	r := NewWebhookReporter(cfg, nil, logger.Noop())
	if r == nil {
		t.Fatal("expected non-nil WebhookReporter")
	}
	if r.client.Timeout != 10*time.Second {
		t.Fatalf("expected default timeout 10s, got %v", r.client.Timeout)
	}
}

func TestNewWebhookReporter_CustomTimeout(t *testing.T) {
	cfg := WebhookConfig{
		URL:     "http://example.com/hook",
		Timeout: 5 * time.Second,
	}
	r := NewWebhookReporter(cfg, nil, logger.Noop())
	if r == nil {
		t.Fatal("expected non-nil WebhookReporter")
	}
	if r.client.Timeout != 5*time.Second {
		t.Fatalf("expected timeout 5s, got %v", r.client.Timeout)
	}
}

func TestAsAppError_WithAppError(t *testing.T) {
	appErr := New("TEST_ERROR", "something went wrong")
	var target *AppError
	if !asAppError(appErr, &target) {
		t.Fatal("expected asAppError to return true for *AppError")
	}
	if target == nil {
		t.Fatal("expected target to be set")
	}
	if target.Type != "TEST_ERROR" {
		t.Fatalf("expected type TEST_ERROR, got %s", target.Type)
	}
}

func TestAsAppError_WithRegularError(t *testing.T) {
	err := fmt.Errorf("plain error")
	var target *AppError
	if asAppError(err, &target) {
		t.Fatal("expected asAppError to return false for regular error")
	}
}

func TestSendError_EnqueuesToRedis(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	cfg := WebhookConfig{URL: "http://example.com/hook"}
	reporter := NewWebhookReporter(cfg, rdb, logger.Noop())

	appErr := New("QUEUE_TEST", "enqueue me").WithDetails("detail-info")
	if err := reporter.SendError(appErr); err != nil {
		t.Fatalf("SendError returned unexpected error: %v", err)
	}

	items, err := mr.List(redisWebhookPendingKey)
	if err != nil {
		t.Fatalf("failed to read list from miniredis: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item in list, got %d", len(items))
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(items[0]), &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if payload["code"] != "QUEUE_TEST" {
		t.Fatalf("expected code QUEUE_TEST, got %v", payload["code"])
	}
}

func TestSendError_NilError(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	cfg := WebhookConfig{URL: "http://example.com/hook"}
	reporter := NewWebhookReporter(cfg, rdb, logger.Noop())

	// A regular (non-AppError) error is silently ignored, returning nil.
	if err := reporter.SendError(fmt.Errorf("not an AppError")); err != nil {
		t.Fatalf("SendError returned unexpected error for non-AppError: %v", err)
	}

	items, _ := mr.List(redisWebhookPendingKey)
	if len(items) != 0 {
		t.Fatalf("expected 0 items in list, got %d", len(items))
	}
}

func TestDoPost_Success(t *testing.T) {
	var received bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := WebhookConfig{URL: ts.URL, Timeout: 5 * time.Second}
	reporter := NewWebhookReporter(cfg, nil, logger.Noop())

	body := []byte(`{"test":"value"}`)
	if err := reporter.doPost(body); err != nil {
		t.Fatalf("doPost returned unexpected error: %v", err)
	}
	if !received {
		t.Fatal("server did not receive the request")
	}
}

func TestDoPost_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	cfg := WebhookConfig{URL: ts.URL, Timeout: 5 * time.Second}
	reporter := NewWebhookReporter(cfg, nil, logger.Noop())

	body := []byte(`{"test":"value"}`)
	err := reporter.doPost(body)
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

// ---------------------------------------------------------------------------
// WebhookFlusher tests
// ---------------------------------------------------------------------------

func TestNewWebhookFlusher(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	cfg := WebhookConfig{URL: "http://example.com/hook"}
	reporter := NewWebhookReporter(cfg, rdb, logger.Noop())

	flusher := NewWebhookFlusher(reporter, rdb, logger.Noop())
	if flusher == nil {
		t.Fatal("expected non-nil WebhookFlusher")
	}
}

func TestWebhookFlusher_Stop_NilCancel(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	cfg := WebhookConfig{URL: "http://example.com/hook"}
	reporter := NewWebhookReporter(cfg, rdb, logger.Noop())
	flusher := NewWebhookFlusher(reporter, rdb, logger.Noop())

	// Stop without Start should not panic.
	flusher.Stop()
}

func TestWebhookFlusher_StartAndStop(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	cfg := WebhookConfig{URL: "http://example.com/hook"}
	reporter := NewWebhookReporter(cfg, rdb, logger.Noop())
	flusher := NewWebhookFlusher(reporter, rdb, logger.Noop())

	ctx := context.Background()
	flusher.Start(ctx)

	// Give the goroutine a moment to spin up.
	time.Sleep(50 * time.Millisecond)

	flusher.Stop()
}
