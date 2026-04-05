package domain_test

import (
	"testing"

	"gct/internal/context/admin/supporting/errorcode/domain"

	"github.com/google/uuid"
)

func TestErrorCodeID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := domain.NewErrorCodeID()
	if id.IsZero() {
		t.Fatal("newly generated ErrorCodeID should not be zero")
	}

	parsed, err := domain.ParseErrorCodeID(id.String())
	if err != nil {
		t.Fatalf("ParseErrorCodeID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
	if parsed.UUID() != id.UUID() {
		t.Fatalf("UUID() mismatch")
	}
}

func TestParseErrorCodeID_Invalid(t *testing.T) {
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
			if _, err := domain.ParseErrorCodeID(tc.in); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
	}
}

func TestErrorCodeID_IsZero(t *testing.T) {
	t.Parallel()

	var zero domain.ErrorCodeID
	if !zero.IsZero() {
		t.Fatal("zero-valued ErrorCodeID should report IsZero()")
	}

	nonZero := domain.ErrorCodeID(uuid.New())
	if nonZero.IsZero() {
		t.Fatal("non-zero ErrorCodeID should not report IsZero()")
	}
}

func TestErrorCodeID_Distinct(t *testing.T) {
	t.Parallel()

	a := domain.NewErrorCodeID()
	b := domain.NewErrorCodeID()
	if a == b {
		t.Fatal("separately generated IDs should differ")
	}
}
