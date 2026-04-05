package domain_test

import (
	"testing"

	"gct/internal/context/ops/generic/ratelimit/domain"

	"github.com/google/uuid"
)

func TestRateLimitID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := domain.NewRateLimitID()
	if id.IsZero() {
		t.Fatal("newly generated RateLimitID should not be zero")
	}

	parsed, err := domain.ParseRateLimitID(id.String())
	if err != nil {
		t.Fatalf("ParseRateLimitID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
	if parsed.UUID() != id.UUID() {
		t.Fatalf("UUID() mismatch")
	}
}

func TestParseRateLimitID_Invalid(t *testing.T) {
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
			if _, err := domain.ParseRateLimitID(tc.in); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
	}
}

func TestRateLimitID_IsZero(t *testing.T) {
	t.Parallel()

	var zero domain.RateLimitID
	if !zero.IsZero() {
		t.Fatal("zero-valued RateLimitID should report IsZero()")
	}

	nonZero := domain.RateLimitID(uuid.New())
	if nonZero.IsZero() {
		t.Fatal("non-zero RateLimitID should not report IsZero()")
	}
}

func TestRateLimitID_Distinct(t *testing.T) {
	t.Parallel()

	a := domain.NewRateLimitID()
	b := domain.NewRateLimitID()
	if a == b {
		t.Fatal("separately generated IDs should differ")
	}
}
