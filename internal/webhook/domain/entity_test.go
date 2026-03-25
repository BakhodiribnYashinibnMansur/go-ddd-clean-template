package domain_test

import (
	"testing"

	"gct/internal/webhook/domain"
)

func TestNewWebhook(t *testing.T) {
	events := []string{"user.created", "user.deleted"}
	w := domain.NewWebhook("my-hook", "https://example.com/hook", "s3cret", events, true)

	if w.Name() != "my-hook" {
		t.Fatalf("expected name my-hook, got %s", w.Name())
	}
	if w.URL() != "https://example.com/hook" {
		t.Fatalf("expected url https://example.com/hook, got %s", w.URL())
	}
	if w.Secret() != "s3cret" {
		t.Fatalf("expected secret s3cret, got %s", w.Secret())
	}
	if len(w.Events_()) != 2 {
		t.Fatalf("expected 2 events, got %d", len(w.Events_()))
	}
	if !w.Enabled() {
		t.Fatal("expected enabled true")
	}
}

func TestWebhook_Trigger(t *testing.T) {
	w := domain.NewWebhook("hook", "https://example.com", "secret", nil, true)

	w.Trigger()
	if len(w.Events()) != 1 {
		t.Fatalf("expected 1 domain event, got %d", len(w.Events()))
	}
	if w.Events()[0].EventName() != "webhook.triggered" {
		t.Fatalf("expected event webhook.triggered, got %s", w.Events()[0].EventName())
	}
}
