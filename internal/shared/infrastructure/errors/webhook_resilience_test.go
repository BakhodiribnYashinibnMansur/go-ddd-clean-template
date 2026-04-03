package errors

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gct/internal/shared/infrastructure/logger"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// TestSendError_RedisDown verifies that SendError does not panic and returns
// nil when Redis is completely unavailable (graceful degradation).
func TestSendError_RedisDown(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })

	cfg := WebhookConfig{URL: "http://example.com/hook"}
	reporter := NewWebhookReporter(cfg, rdb, logger.Noop())

	// Shut down Redis.
	mr.Close()

	appErr := New("REDIS_DOWN_TEST", "should not panic").WithDetails("detail")

	// Must not panic.
	err := reporter.SendError(appErr)
	if err != nil {
		t.Fatalf("SendError() returned error when Redis is down: %v; expected nil (graceful degradation)", err)
	}
}

// TestDoPost_Timeout verifies that doPost returns an error when the remote
// server takes longer than the configured timeout.
func TestDoPost_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := WebhookConfig{
		URL:     ts.URL,
		Timeout: 100 * time.Millisecond,
	}
	reporter := NewWebhookReporter(cfg, nil, logger.Noop())

	err := reporter.doPost([]byte(`{"test":"timeout"}`))
	if err == nil {
		t.Fatal("doPost() expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "send webhook") {
		t.Errorf("expected error to contain 'send webhook', got: %v", err)
	}
}

// TestDoPost_ConnectionRefused verifies that doPost returns an error when the
// URL points to a server that is not listening.
func TestDoPost_ConnectionRefused(t *testing.T) {
	cfg := WebhookConfig{
		URL:     "http://127.0.0.1:1", // port 1 — nothing listening
		Timeout: 1 * time.Second,
	}
	reporter := NewWebhookReporter(cfg, nil, logger.Noop())

	err := reporter.doPost([]byte(`{"test":"refused"}`))
	if err == nil {
		t.Fatal("doPost() expected connection refused error, got nil")
	}
	if !strings.Contains(err.Error(), "send webhook") {
		t.Errorf("expected error to contain 'send webhook', got: %v", err)
	}
}
