package errorx_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	apperrors "gct/internal/kernel/infrastructure/errorx"
)

func TestWebhookReporter_EnqueuesError(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	reporter := apperrors.NewWebhookReporter(apperrors.WebhookConfig{
		URL: "http://unused",
	}, rdb, nil)

	err := apperrors.New(apperrors.ErrRepoConnection, "")
	if sendErr := reporter.SendError(err); sendErr != nil {
		t.Fatalf("SendError returned error: %v", sendErr)
	}

	entries, listErr := rdb.LRange(context.Background(), "webhook:pending", 0, -1).Result()
	if listErr != nil {
		t.Fatalf("LRange failed: %v", listErr)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 queued entry, got %d", len(entries))
	}

	var payload map[string]any
	if jsonErr := json.Unmarshal([]byte(entries[0]), &payload); jsonErr != nil {
		t.Fatalf("unmarshal failed: %v", jsonErr)
	}
	if payload["code"] != apperrors.ErrRepoConnection {
		t.Fatalf("expected code %s, got %v", apperrors.ErrRepoConnection, payload["code"])
	}
}

func TestWebhookReporter_NilRedisDropsGracefully(t *testing.T) {
	reporter := apperrors.NewWebhookReporter(apperrors.WebhookConfig{
		URL: "http://unused",
	}, nil, nil)

	err := apperrors.New(apperrors.ErrRepoConnection, "")
	if sendErr := reporter.SendError(err); sendErr != nil {
		t.Fatalf("SendError should return nil when Redis is nil, got: %v", sendErr)
	}
}

func TestWebhookReporter_SkipsNonAppError(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	reporter := apperrors.NewWebhookReporter(apperrors.WebhookConfig{
		URL: server.URL,
	}, nil, nil)

	reporter.SendError(nil)
	if called {
		t.Fatal("should not call webhook for nil error")
	}
}
