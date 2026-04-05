package domain_test

import (
	"testing"

	"gct/internal/context/content/generic/file/domain"

	"github.com/google/uuid"
)

func TestFileID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := domain.NewFileID()
	if id.IsZero() {
		t.Fatal("newly generated FileID should not be zero")
	}

	parsed, err := domain.ParseFileID(id.String())
	if err != nil {
		t.Fatalf("ParseFileID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
	if parsed.UUID() != id.UUID() {
		t.Fatalf("UUID() mismatch")
	}
}

func TestParseFileID_Invalid(t *testing.T) {
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
			if _, err := domain.ParseFileID(tc.in); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
	}
}

func TestFileID_IsZero(t *testing.T) {
	t.Parallel()

	var zero domain.FileID
	if !zero.IsZero() {
		t.Fatal("zero-valued FileID should report IsZero()")
	}

	nonZero := domain.FileID(uuid.New())
	if nonZero.IsZero() {
		t.Fatal("non-zero FileID should not report IsZero()")
	}
}

func TestFileID_Distinct(t *testing.T) {
	t.Parallel()

	a := domain.NewFileID()
	b := domain.NewFileID()
	if a == b {
		t.Fatal("separately generated IDs should differ")
	}
}
