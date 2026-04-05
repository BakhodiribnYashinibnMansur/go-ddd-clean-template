package domain_test

import (
	"testing"

	"gct/internal/context/iam/audit/domain"

	"github.com/google/uuid"
)

func TestAuditLogID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := domain.NewAuditLogID()
	if id.IsZero() {
		t.Fatal("newly generated AuditLogID should not be zero")
	}

	parsed, err := domain.ParseAuditLogID(id.String())
	if err != nil {
		t.Fatalf("ParseAuditLogID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
	if parsed.UUID() != id.UUID() {
		t.Fatalf("UUID() mismatch")
	}
}

func TestParseAuditLogID_Invalid(t *testing.T) {
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
			if _, err := domain.ParseAuditLogID(tc.in); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
	}
}

func TestAuditLogID_IsZero(t *testing.T) {
	t.Parallel()

	var zero domain.AuditLogID
	if !zero.IsZero() {
		t.Fatal("zero-valued AuditLogID should report IsZero()")
	}

	nonZero := domain.AuditLogID(uuid.New())
	if nonZero.IsZero() {
		t.Fatal("non-zero AuditLogID should not report IsZero()")
	}
}

func TestEndpointHistoryID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := domain.NewEndpointHistoryID()
	if id.IsZero() {
		t.Fatal("newly generated EndpointHistoryID should not be zero")
	}

	parsed, err := domain.ParseEndpointHistoryID(id.String())
	if err != nil {
		t.Fatalf("ParseEndpointHistoryID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
}

func TestParseEndpointHistoryID_Invalid(t *testing.T) {
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
			if _, err := domain.ParseEndpointHistoryID(tc.in); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
	}
}

func TestEndpointHistoryID_IsZero(t *testing.T) {
	t.Parallel()

	var zero domain.EndpointHistoryID
	if !zero.IsZero() {
		t.Fatal("zero-valued EndpointHistoryID should report IsZero()")
	}
}

func TestAuditLogID_DistinctFromEndpointHistoryID(t *testing.T) {
	t.Parallel()

	// Compile-time safety: the following line would fail to compile:
	//   var alid domain.AuditLogID = domain.NewEndpointHistoryID()
	alid := domain.NewAuditLogID()
	ehid := domain.NewEndpointHistoryID()
	if alid.String() == ehid.String() {
		t.Fatal("separately generated IDs should differ")
	}
}
