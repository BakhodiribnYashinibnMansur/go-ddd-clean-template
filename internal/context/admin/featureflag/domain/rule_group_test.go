package domain_test

import (
	"testing"

	"gct/internal/context/admin/featureflag/domain"

	"github.com/google/uuid"
)

func TestRuleGroup_MatchAll_AllConditionsTrue(t *testing.T) {
	t.Parallel()

	rg := domain.NewRuleGroup(uuid.New(), "beta-users", "true", 1)
	rg.AddCondition(domain.NewCondition("country", "eq", "US"))
	rg.AddCondition(domain.NewCondition("age", "gte", "18"))

	attrs := map[string]string{"country": "US", "age": "25"}
	if !rg.MatchAll(attrs) {
		t.Fatal("expected all conditions to match")
	}
}

func TestRuleGroup_MatchAll_OneConditionFalse(t *testing.T) {
	t.Parallel()

	rg := domain.NewRuleGroup(uuid.New(), "beta-users", "true", 1)
	rg.AddCondition(domain.NewCondition("country", "eq", "US"))
	rg.AddCondition(domain.NewCondition("age", "gte", "18"))

	attrs := map[string]string{"country": "US", "age": "16"}
	if rg.MatchAll(attrs) {
		t.Fatal("expected match to fail when one condition is false")
	}
}

func TestRuleGroup_MatchAll_MissingAttribute(t *testing.T) {
	t.Parallel()

	rg := domain.NewRuleGroup(uuid.New(), "beta-users", "true", 1)
	rg.AddCondition(domain.NewCondition("country", "eq", "US"))

	attrs := map[string]string{"age": "25"}
	if rg.MatchAll(attrs) {
		t.Fatal("expected match to fail when attribute is missing")
	}
}

func TestRuleGroup_MatchAll_NoConditions(t *testing.T) {
	t.Parallel()

	rg := domain.NewRuleGroup(uuid.New(), "empty-group", "true", 1)

	attrs := map[string]string{"country": "US"}
	if rg.MatchAll(attrs) {
		t.Fatal("expected no match when there are no conditions")
	}
}
