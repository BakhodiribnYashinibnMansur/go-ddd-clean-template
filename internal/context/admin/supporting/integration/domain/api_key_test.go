package domain_test

import (
	"errors"
	"strings"
	"testing"

	"gct/internal/context/admin/supporting/integration/domain"
)

func TestNewAPIKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid 16 chars", input: "abcdefghijklmnop"},
		{name: "valid 32 chars", input: "abcdefghijklmnopqrstuvwxyz012345"},
		{name: "valid long key", input: strings.Repeat("x", 64)},
		{name: "too short 15 chars", input: "abcdefghijklmno", wantErr: true},
		{name: "too short 1 char", input: "a", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := domain.NewAPIKey(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil", tt.input)
				}
				if !errors.Is(err, domain.ErrInvalidAPIKey) {
					t.Fatalf("expected ErrInvalidAPIKey, got %v", err)
				}
				if !got.IsZero() {
					t.Fatalf("expected zero value on error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Reveal() != tt.input {
				t.Errorf("Reveal() = %q, want %q", got.Reveal(), tt.input)
			}
			if got.IsZero() {
				t.Errorf("IsZero() = true, want false")
			}
		})
	}
}

func TestAPIKey_Redaction(t *testing.T) {
	t.Parallel()
	raw := "super-secret-api-key-12345"
	k, err := domain.NewAPIKey(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if k.String() == k.Reveal() {
		t.Fatalf("String() must not equal Reveal()")
	}
	if k.String() != "[REDACTED]" {
		t.Errorf("String() = %q, want %q", k.String(), "[REDACTED]")
	}
	if strings.Contains(k.String(), raw) {
		t.Errorf("String() leaks raw secret: %q", k.String())
	}
}

func TestAPIKey_Equal(t *testing.T) {
	t.Parallel()
	a, _ := domain.NewAPIKey("abcdefghijklmnop")
	b, _ := domain.NewAPIKey("abcdefghijklmnop")
	c, _ := domain.NewAPIKey("zzzzzzzzzzzzzzzz")
	if !a.Equal(b) {
		t.Errorf("expected a.Equal(b) = true")
	}
	if a.Equal(c) {
		t.Errorf("expected a.Equal(c) = false")
	}
}

func TestAPIKey_IsZero(t *testing.T) {
	t.Parallel()
	var k domain.APIKey
	if !k.IsZero() {
		t.Errorf("zero value IsZero() = false, want true")
	}
}
