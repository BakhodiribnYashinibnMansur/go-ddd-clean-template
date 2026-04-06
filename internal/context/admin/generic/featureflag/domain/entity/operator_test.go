package entity_test

import (
	"testing"

	"gct/internal/context/admin/generic/featureflag/domain/entity"
)

func TestIsValidOperator_ValidOperators(t *testing.T) {
	t.Parallel()

	validOps := []string{
		entity.OpEq, entity.OpNotEq,
		entity.OpIn, entity.OpNotIn,
		entity.OpGt, entity.OpGte,
		entity.OpLt, entity.OpLte,
		entity.OpContains,
	}

	for _, op := range validOps {
		if !entity.IsValidOperator(op) {
			t.Errorf("expected operator %q to be valid", op)
		}
	}
}

func TestIsValidOperator_InvalidOperators(t *testing.T) {
	t.Parallel()

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
		if entity.IsValidOperator(op) {
			t.Errorf("expected operator %q to be invalid", op)
		}
	}
}

func TestValidOperatorsMap_Count(t *testing.T) {
	t.Parallel()

	expected := 9
	if len(entity.ValidOperators) != expected {
		t.Errorf("expected %d valid operators, got %d", expected, len(entity.ValidOperators))
	}
}
