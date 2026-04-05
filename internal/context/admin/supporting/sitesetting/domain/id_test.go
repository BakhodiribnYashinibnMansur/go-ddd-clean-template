package domain_test

import (
	"testing"

	"gct/internal/context/admin/supporting/sitesetting/domain"

	"github.com/google/uuid"
)

func TestSiteSettingID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := domain.NewSiteSettingID()
	if id.IsZero() {
		t.Fatal("newly generated SiteSettingID should not be zero")
	}

	parsed, err := domain.ParseSiteSettingID(id.String())
	if err != nil {
		t.Fatalf("ParseSiteSettingID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
	if parsed.UUID() != id.UUID() {
		t.Fatalf("UUID() mismatch")
	}
}

func TestParseSiteSettingID_Invalid(t *testing.T) {
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
			if _, err := domain.ParseSiteSettingID(tc.in); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
	}
}

func TestSiteSettingID_IsZero(t *testing.T) {
	t.Parallel()

	var zero domain.SiteSettingID
	if !zero.IsZero() {
		t.Fatal("zero-valued SiteSettingID should report IsZero()")
	}

	nonZero := domain.SiteSettingID(uuid.New())
	if nonZero.IsZero() {
		t.Fatal("non-zero SiteSettingID should not report IsZero()")
	}
}

func TestSiteSettingID_Distinct(t *testing.T) {
	t.Parallel()

	a := domain.NewSiteSettingID()
	b := domain.NewSiteSettingID()
	if a == b {
		t.Fatal("separately generated IDs should differ")
	}
}
