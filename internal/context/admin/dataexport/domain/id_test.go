package domain_test

import (
	"testing"

	"gct/internal/context/admin/dataexport/domain"

	"github.com/google/uuid"
)

func TestDataExportID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := domain.NewDataExportID()
	if id.IsZero() {
		t.Fatal("newly generated DataExportID should not be zero")
	}

	parsed, err := domain.ParseDataExportID(id.String())
	if err != nil {
		t.Fatalf("ParseDataExportID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
	if parsed.UUID() != id.UUID() {
		t.Fatalf("UUID() mismatch")
	}
}

func TestParseDataExportID_Invalid(t *testing.T) {
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
			if _, err := domain.ParseDataExportID(tc.in); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
	}
}

func TestDataExportID_IsZero(t *testing.T) {
	t.Parallel()

	var zero domain.DataExportID
	if !zero.IsZero() {
		t.Fatal("zero-valued DataExportID should report IsZero()")
	}

	nonZero := domain.DataExportID(uuid.New())
	if nonZero.IsZero() {
		t.Fatal("non-zero DataExportID should not report IsZero()")
	}
}

func TestDataExportID_Distinct(t *testing.T) {
	t.Parallel()

	a := domain.NewDataExportID()
	b := domain.NewDataExportID()
	if a == b {
		t.Fatal("separately generated IDs should differ")
	}
}
