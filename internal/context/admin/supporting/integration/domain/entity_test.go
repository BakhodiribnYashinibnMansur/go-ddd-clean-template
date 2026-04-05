package domain_test

import (
	"testing"
	"time"

	"gct/internal/context/admin/supporting/integration/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIntegration(t *testing.T) {
	t.Parallel()

	config := map[string]string{"timeout": "30"}
	i, _ := domain.NewIntegration("Stripe", "payment", "sk_test_123", "https://hooks.example.com", true, config)

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

func TestIntegration_SetJWT_RejectsInvalidTTLs(t *testing.T) {
	t.Parallel()
	i, _ := domain.NewIntegration("Stripe", "payment", "sk", "https://e", true, nil)

	err := i.SetJWT([]byte("h"), 0, 1*time.Hour, "pem", "kid", domain.BindingModeWarn, 0)
	require.Error(t, err)

	err = i.SetJWT([]byte("h"), 1*time.Hour, -1, "pem", "kid", domain.BindingModeWarn, 0)
	require.Error(t, err)
}

func TestIntegration_SetJWT_RejectsInvalidMode(t *testing.T) {
	t.Parallel()
	i, _ := domain.NewIntegration("Stripe", "payment", "sk", "https://e", true, nil)

	err := i.SetJWT([]byte("h"), 1*time.Hour, 24*time.Hour, "pem", "kid", "bogus", 0)
	require.Error(t, err)
}

func TestIntegration_SetJWT_AllValidModes(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{domain.BindingModeOff, domain.BindingModeWarn, domain.BindingModeStrict} {
		i, _ := domain.NewIntegration("Stripe", "payment", "sk", "https://e", true, nil)
		err := i.SetJWT([]byte("h"), 1*time.Hour, 24*time.Hour, "pem", "kid", mode, 5)
		require.NoError(t, err, "mode %q should be accepted", mode)
		assert.Equal(t, mode, i.JWTBindingMode())
		assert.Equal(t, 5, i.JWTMaxSessions())
		assert.Equal(t, 1*time.Hour, i.JWTAccessTTL())
		assert.Equal(t, 24*time.Hour, i.JWTRefreshTTL())
	}
}

func TestIntegration_RotateJWTKey(t *testing.T) {
	t.Parallel()
	i, _ := domain.NewIntegration("Stripe", "payment", "sk", "https://e", true, nil)
	_ = i.SetJWT([]byte("h"), 1*time.Hour, 24*time.Hour, "first-pem", "kid-1", domain.BindingModeWarn, 0)

	assert.Nil(t, i.JWTRotatedAt())
	assert.Empty(t, i.JWTPreviousPublicKeyPEM())
	assert.Empty(t, i.JWTPreviousKeyID())

	i.RotateJWTKey("second-pem", "kid-2")

	assert.Equal(t, "second-pem", i.JWTPublicKeyPEM())
	assert.Equal(t, "kid-2", i.JWTKeyID())
	assert.Equal(t, "first-pem", i.JWTPreviousPublicKeyPEM())
	assert.Equal(t, "kid-1", i.JWTPreviousKeyID())
	require.NotNil(t, i.JWTRotatedAt())
	assert.WithinDuration(t, time.Now(), *i.JWTRotatedAt(), 2*time.Second)
}
