package entity_test

import (
	"testing"

	"gct/internal/context/content/supporting/announcement/domain/entity"

	"github.com/google/uuid"
)

func TestAnnouncementID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := entity.NewAnnouncementID()
	if id.IsZero() {
		t.Fatal("newly generated AnnouncementID should not be zero")
	}

	parsed, err := entity.ParseAnnouncementID(id.String())
	if err != nil {
		t.Fatalf("ParseAnnouncementID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
	if parsed.UUID() != id.UUID() {
		t.Fatalf("UUID() mismatch")
	}
}

func TestParseAnnouncementID_Invalid(t *testing.T) {
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
			if _, err := entity.ParseAnnouncementID(tc.in); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
	}
}

func TestAnnouncementID_IsZero(t *testing.T) {
	t.Parallel()

	var zero entity.AnnouncementID
	if !zero.IsZero() {
		t.Fatal("zero-valued AnnouncementID should report IsZero()")
	}

	nonZero := entity.AnnouncementID(uuid.New())
	if nonZero.IsZero() {
		t.Fatal("non-zero AnnouncementID should not report IsZero()")
	}
}

func TestAnnouncementID_Distinct(t *testing.T) {
	t.Parallel()

	a := entity.NewAnnouncementID()
	b := entity.NewAnnouncementID()
	if a == b {
		t.Fatal("separately generated IDs should differ")
	}
}
