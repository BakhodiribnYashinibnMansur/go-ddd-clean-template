package domain_test

import (
	"testing"

	"gct/internal/featureflag/domain"
)

func TestIsValidOperator_ValidOperators(t *testing.T) {
	validOps := []string{
		domain.OpEq, domain.OpNotEq,
		domain.OpIn, domain.OpNotIn,
		domain.OpGt, domain.OpGte,
		domain.OpLt, domain.OpLte,
		domain.OpContains,
	}

	for _, op := range validOps {
		if !domain.IsValidOperator(op) {
			t.Errorf("expected operator %q to be valid", op)
		}
	}
}

func TestIsValidOperator_InvalidOperators(t *testing.T) {
	invalidOps := []string{
		"",
		"invalid",
		"equals",
		"EQ",
		"GT",
		"like",
		"between",
		"regex",
	}

	for _, op := range invalidOps {
		if domain.IsValidOperator(op) {
			t.Errorf("expected operator %q to be invalid", op)
		}
	}
}

func TestValidOperatorsMap_Count(t *testing.T) {
	expected := 9
	if len(domain.ValidOperators) != expected {
		t.Errorf("expected %d valid operators, got %d", expected, len(domain.ValidOperators))
	}
}
