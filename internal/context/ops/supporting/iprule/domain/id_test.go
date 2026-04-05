package domain_test

import (
	"testing"

	"gct/internal/context/ops/supporting/iprule/domain"

	"github.com/google/uuid"
)

func TestIPRuleID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := domain.NewIPRuleID()
	if id.IsZero() {
		t.Fatal("newly generated IPRuleID should not be zero")
	}

	parsed, err := domain.ParseIPRuleID(id.String())
	if err != nil {
		t.Fatalf("ParseIPRuleID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
	if parsed.UUID() != id.UUID() {
		t.Fatalf("UUID() mismatch")
	}
}

func TestParseIPRuleID_Invalid(t *testing.T) {
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
			if _, err := domain.ParseIPRuleID(tc.in); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
	}
}

func TestIPRuleID_IsZero(t *testing.T) {
	t.Parallel()

	var zero domain.IPRuleID
	if !zero.IsZero() {
		t.Fatal("zero-valued IPRuleID should report IsZero()")
	}

	nonZero := domain.IPRuleID(uuid.New())
	if nonZero.IsZero() {
		t.Fatal("non-zero IPRuleID should not report IsZero()")
	}
}

func TestIPRuleID_Distinct(t *testing.T) {
	t.Parallel()

	a := domain.NewIPRuleID()
	b := domain.NewIPRuleID()
	if a == b {
		t.Fatal("separately generated IDs should differ")
	}
}
