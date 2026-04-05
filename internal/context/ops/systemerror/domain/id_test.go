package domain_test

import (
	"testing"

	"gct/internal/context/ops/systemerror/domain"

	"github.com/google/uuid"
)

func TestSystemErrorID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := domain.NewSystemErrorID()
	if id.IsZero() {
		t.Fatal("newly generated SystemErrorID should not be zero")
	}

	parsed, err := domain.ParseSystemErrorID(id.String())
	if err != nil {
		t.Fatalf("ParseSystemErrorID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
	if parsed.UUID() != id.UUID() {
		t.Fatalf("UUID() mismatch")
	}
}

func TestParseSystemErrorID_Invalid(t *testing.T) {
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
			if _, err := domain.ParseSystemErrorID(tc.in); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
	}
}

func TestSystemErrorID_IsZero(t *testing.T) {
	t.Parallel()

	var zero domain.SystemErrorID
	if !zero.IsZero() {
		t.Fatal("zero-valued SystemErrorID should report IsZero()")
	}

	nonZero := domain.SystemErrorID(uuid.New())
	if nonZero.IsZero() {
		t.Fatal("non-zero SystemErrorID should not report IsZero()")
	}
}

func TestSystemErrorID_Distinct(t *testing.T) {
	t.Parallel()

	a := domain.NewSystemErrorID()
	b := domain.NewSystemErrorID()
	if a == b {
		t.Fatal("separately generated IDs should differ")
	}
}
