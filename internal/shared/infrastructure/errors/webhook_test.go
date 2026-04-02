package errors_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "gct/internal/shared/infrastructure/errors"
)

func TestWebhookReporter_SendsError(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	reporter := apperrors.NewWebhookReporter(apperrors.WebhookConfig{
		URL: server.URL,
	})

	err := apperrors.New(apperrors.ErrRepoConnection, "")
	reporter.SendError(err)

	if received["code"] != apperrors.ErrRepoConnection {
		t.Fatalf("expected code %s, got %v", apperrors.ErrRepoConnection, received["code"])
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
	})

	reporter.SendError(nil)
	if called {
		t.Fatal("should not call webhook for nil error")
	}
}
