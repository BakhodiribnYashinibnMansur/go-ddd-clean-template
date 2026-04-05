package domain_test

import (
	"testing"

	"gct/internal/context/content/translation/domain"

	"github.com/google/uuid"
)

func TestTranslationID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := domain.NewTranslationID()
	if id.IsZero() {
		t.Fatal("newly generated TranslationID should not be zero")
	}

	parsed, err := domain.ParseTranslationID(id.String())
	if err != nil {
		t.Fatalf("ParseTranslationID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
	if parsed.UUID() != id.UUID() {
		t.Fatalf("UUID() mismatch")
	}
}

func TestParseTranslationID_Invalid(t *testing.T) {
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
			if _, err := domain.ParseTranslationID(tc.in); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
	}
}

func TestTranslationID_IsZero(t *testing.T) {
	t.Parallel()

	var zero domain.TranslationID
	if !zero.IsZero() {
		t.Fatal("zero-valued TranslationID should report IsZero()")
	}

	nonZero := domain.TranslationID(uuid.New())
	if nonZero.IsZero() {
		t.Fatal("non-zero TranslationID should not report IsZero()")
	}
}

func TestTranslationID_Distinct(t *testing.T) {
	t.Parallel()

	a := domain.NewTranslationID()
	b := domain.NewTranslationID()
	if a == b {
		t.Fatal("separately generated IDs should differ")
	}
}
