package domain_test

import (
	"testing"

	"gct/internal/context/admin/integration/domain"

	"github.com/google/uuid"
)

func TestIntegrationID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := domain.NewIntegrationID()
	if id.IsZero() {
		t.Fatal("newly generated IntegrationID should not be zero")
	}

	parsed, err := domain.ParseIntegrationID(id.String())
	if err != nil {
		t.Fatalf("ParseIntegrationID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
	if parsed.UUID() != id.UUID() {
		t.Fatalf("UUID() mismatch")
	}
}

func TestParseIntegrationID_Invalid(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, in string }{
		{"empty", ""},
		{"garbage", "not-a-uuid"},
		{"truncated", "123e4567-e89b-12d3-a456"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := domain.ParseIntegrationID(tc.in); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
	}
}

func TestIntegrationID_IsZero(t *testing.T) {
	t.Parallel()

	var zero domain.IntegrationID
	if !zero.IsZero() {
		t.Fatal("zero-valued IntegrationID should report IsZero()")
	}

	nonZero := domain.IntegrationID(uuid.New())
	if nonZero.IsZero() {
		t.Fatal("non-zero IntegrationID should not report IsZero()")
	}
}

func TestIntegrationID_Distinct(t *testing.T) {
	t.Parallel()

	a := domain.NewIntegrationID()
	b := domain.NewIntegrationID()
	if a == b {
		t.Fatal("separately generated IDs should differ")
	}
}
