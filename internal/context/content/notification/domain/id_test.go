package domain_test

import (
	"testing"

	"gct/internal/context/content/notification/domain"

	"github.com/google/uuid"
)

func TestNotificationID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := domain.NewNotificationID()
	if id.IsZero() {
		t.Fatal("newly generated NotificationID should not be zero")
	}

	parsed, err := domain.ParseNotificationID(id.String())
	if err != nil {
		t.Fatalf("ParseNotificationID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
	if parsed.UUID() != id.UUID() {
		t.Fatalf("UUID() mismatch")
	}
}

func TestParseNotificationID_Invalid(t *testing.T) {
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
			if _, err := domain.ParseNotificationID(tc.in); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
	}
}

func TestNotificationID_IsZero(t *testing.T) {
	t.Parallel()

	var zero domain.NotificationID
	if !zero.IsZero() {
		t.Fatal("zero-valued NotificationID should report IsZero()")
	}

	nonZero := domain.NotificationID(uuid.New())
	if nonZero.IsZero() {
		t.Fatal("non-zero NotificationID should not report IsZero()")
	}
}

func TestNotificationID_Distinct(t *testing.T) {
	t.Parallel()

	a := domain.NewNotificationID()
	b := domain.NewNotificationID()
	if a == b {
		t.Fatal("separately generated IDs should differ")
	}
}
