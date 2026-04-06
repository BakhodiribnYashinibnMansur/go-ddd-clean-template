package entity_test

import (
	"testing"

	"gct/internal/context/iam/generic/session/domain/entity"

	"github.com/google/uuid"
)

func TestSessionID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := entity.NewSessionID()
	if id.IsZero() {
		t.Fatal("newly generated SessionID should not be zero")
	}

	parsed, err := entity.ParseSessionID(id.String())
	if err != nil {
		t.Fatalf("ParseSessionID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
	if parsed.UUID() != id.UUID() {
		t.Fatalf("UUID() mismatch")
	}
}

func TestParseSessionID_Invalid(t *testing.T) {
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
			if _, err := entity.ParseSessionID(tc.in); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
	}
}

func TestSessionID_IsZero(t *testing.T) {
	t.Parallel()

	var zero entity.SessionID
	if !zero.IsZero() {
		t.Fatal("zero-valued SessionID should report IsZero()")
	}

	nonZero := entity.SessionID(uuid.New())
	if nonZero.IsZero() {
		t.Fatal("non-zero SessionID should not report IsZero()")
	}
}

func TestUserID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := entity.NewUserID()
	if id.IsZero() {
		t.Fatal("newly generated UserID should not be zero")
	}

	parsed, err := entity.ParseUserID(id.String())
	if err != nil {
		t.Fatalf("ParseUserID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
}

func TestParseUserID_Invalid(t *testing.T) {
	t.Parallel()

	if _, err := entity.ParseUserID("not-a-uuid"); err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestUserID_IsZero(t *testing.T) {
	t.Parallel()

	var zero entity.UserID
	if !zero.IsZero() {
		t.Fatal("zero-valued UserID should report IsZero()")
	}

	nonZero := entity.UserID(uuid.New())
	if nonZero.IsZero() {
		t.Fatal("non-zero UserID should not report IsZero()")
	}
}

func TestSessionID_DistinctFromUserID(t *testing.T) {
	t.Parallel()

	// Compile-time safety: the following would not compile:
	//   var uid entity.UserID = entity.NewSessionID()
	sid := entity.NewSessionID()
	uid := entity.NewUserID()
	if sid.String() == uid.String() {
		t.Fatal("separately generated IDs should differ")
	}
}
