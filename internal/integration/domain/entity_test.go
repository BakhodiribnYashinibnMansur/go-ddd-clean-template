package domain_test

import (
	"testing"

	"gct/internal/integration/domain"
)

func TestNewIntegration(t *testing.T) {
	config := map[string]string{"timeout": "30"}
	i := domain.NewIntegration("Stripe", "payment", "sk_test_123", "https://hooks.example.com", true, config)

	if i.Name() != "Stripe" {
		t.Fatalf("expected name Stripe, got %s", i.Name())
	}
	if i.Type() != "payment" {
		t.Fatalf("expected type payment, got %s", i.Type())
	}
	if i.APIKey() != "sk_test_123" {
		t.Fatalf("expected apiKey sk_test_123, got %s", i.APIKey())
	}
	if i.WebhookURL() != "https://hooks.example.com" {
		t.Fatalf("expected webhookURL https://hooks.example.com, got %s", i.WebhookURL())
	}
	if !i.Enabled() {
		t.Fatal("expected enabled true")
	}
	if i.Config()["timeout"] != "30" {
		t.Fatalf("expected config timeout 30, got %v", i.Config()["timeout"])
	}
	if len(i.Events()) != 1 {
		t.Fatalf("expected 1 event, got %d", len(i.Events()))
	}
	if i.Events()[0].EventName() != "integration.connected" {
		t.Fatalf("expected event integration.connected, got %s", i.Events()[0].EventName())
	}
}
