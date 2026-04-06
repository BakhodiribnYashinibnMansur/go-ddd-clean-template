package entity_test

import (
	"testing"

	"gct/internal/context/iam/generic/user/domain/entity"

	"github.com/google/uuid"
)

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
	if parsed.UUID() != id.UUID() {
		t.Fatalf("UUID() mismatch")
	}
}

func TestParseUserID_Invalid(t *testing.T) {
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
			if _, err := entity.ParseUserID(tc.in); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
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

func TestSessionID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := entity.NewSessionID()
	parsed, err := entity.ParseSessionID(id.String())
	if err != nil {
		t.Fatalf("ParseSessionID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
}

func TestSessionID_DistinctFromUserID(t *testing.T) {
	t.Parallel()

	// This test's main purpose is to document compile-time safety:
	// the following line would fail to compile:
	//   var uid entity.UserID = entity.NewSessionID()
	// which is precisely the value of typed IDs.
	sid := entity.NewSessionID()
	uid := entity.NewUserID()
	if sid.String() == uid.String() {
		t.Fatal("separately generated IDs should differ")
	}
}
