package entity_test

import (
	"testing"

	"gct/internal/context/admin/generic/featureflag/domain/entity"

	"github.com/google/uuid"
)

func TestFeatureFlagID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := entity.NewFeatureFlagID()
	if id.IsZero() {
		t.Fatal("newly generated FeatureFlagID should not be zero")
	}

	parsed, err := entity.ParseFeatureFlagID(id.String())
	if err != nil {
		t.Fatalf("ParseFeatureFlagID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
	if parsed.UUID() != id.UUID() {
		t.Fatalf("UUID() mismatch")
	}
}

func TestParseFeatureFlagID_Invalid(t *testing.T) {
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
			if _, err := entity.ParseFeatureFlagID(tc.in); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
	}
}

func TestFeatureFlagID_IsZero(t *testing.T) {
	t.Parallel()

	var zero entity.FeatureFlagID
	if !zero.IsZero() {
		t.Fatal("zero-valued FeatureFlagID should report IsZero()")
	}

	nonZero := entity.FeatureFlagID(uuid.New())
	if nonZero.IsZero() {
		t.Fatal("non-zero FeatureFlagID should not report IsZero()")
	}
}

func TestFeatureFlagID_Distinct(t *testing.T) {
	t.Parallel()

	a := entity.NewFeatureFlagID()
	b := entity.NewFeatureFlagID()
	if a == b {
		t.Fatal("separately generated IDs should differ")
	}
}

func TestRuleGroupID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := entity.NewRuleGroupID()
	if id.IsZero() {
		t.Fatal("newly generated RuleGroupID should not be zero")
	}

	parsed, err := entity.ParseRuleGroupID(id.String())
	if err != nil {
		t.Fatalf("ParseRuleGroupID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
	if parsed.UUID() != id.UUID() {
		t.Fatalf("UUID() mismatch")
	}
}

func TestParseRuleGroupID_Invalid(t *testing.T) {
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
			if _, err := entity.ParseRuleGroupID(tc.in); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
	}
}

func TestRuleGroupID_IsZero(t *testing.T) {
	t.Parallel()

	var zero entity.RuleGroupID
	if !zero.IsZero() {
		t.Fatal("zero-valued RuleGroupID should report IsZero()")
	}

	nonZero := entity.RuleGroupID(uuid.New())
	if nonZero.IsZero() {
		t.Fatal("non-zero RuleGroupID should not report IsZero()")
	}
}

func TestRuleGroupID_Distinct(t *testing.T) {
	t.Parallel()

	a := entity.NewRuleGroupID()
	b := entity.NewRuleGroupID()
	if a == b {
		t.Fatal("separately generated IDs should differ")
	}
}
