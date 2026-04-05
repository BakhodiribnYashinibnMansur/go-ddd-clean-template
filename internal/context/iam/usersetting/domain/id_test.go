package domain_test

import (
	"testing"

	"gct/internal/context/iam/usersetting/domain"

	"github.com/google/uuid"
)

func TestUserSettingID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := domain.NewUserSettingID()
	if id.IsZero() {
		t.Fatal("newly generated UserSettingID should not be zero")
	}

	parsed, err := domain.ParseUserSettingID(id.String())
	if err != nil {
		t.Fatalf("ParseUserSettingID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
	if parsed.UUID() != id.UUID() {
		t.Fatalf("UUID() mismatch")
	}
}

func TestParseUserSettingID_Invalid(t *testing.T) {
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
			if _, err := domain.ParseUserSettingID(tc.in); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
	}
}

func TestUserSettingID_IsZero(t *testing.T) {
	t.Parallel()

	var zero domain.UserSettingID
	if !zero.IsZero() {
		t.Fatal("zero-valued UserSettingID should report IsZero()")
	}

	nonZero := domain.UserSettingID(uuid.New())
	if nonZero.IsZero() {
		t.Fatal("non-zero UserSettingID should not report IsZero()")
	}
}

func TestUserSettingID_Distinct(t *testing.T) {
	t.Parallel()

	a := domain.NewUserSettingID()
	b := domain.NewUserSettingID()
	if a == b {
		t.Fatal("separately generated IDs should differ")
	}
}
