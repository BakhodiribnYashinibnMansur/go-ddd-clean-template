package domain_test

import (
	"testing"
	"time"

	"gct/internal/context/admin/featureflag/domain"

	"github.com/google/uuid"
)

func TestNewFeatureFlag(t *testing.T) {
	t.Parallel()

	ff, _ := domain.NewFeatureFlag("dark-mode", "dark_mode", "Enable dark mode", "bool", "false", 50)

	if ff.Name() != "dark-mode" {
		t.Fatalf("expected name dark-mode, got %s", ff.Name())
	}
	if ff.Key() != "dark_mode" {
		t.Fatalf("expected key dark_mode, got %s", ff.Key())
	}
	if ff.Description() != "Enable dark mode" {
		t.Fatalf("expected description 'Enable dark mode', got %s", ff.Description())
	}
	if ff.FlagType() != "bool" {
		t.Fatalf("expected flagType bool, got %s", ff.FlagType())
	}
	if ff.DefaultValue() != "false" {
		t.Fatalf("expected defaultValue false, got %s", ff.DefaultValue())
	}
	if ff.RolloutPercentage() != 50 {
		t.Fatalf("expected rollout 50, got %d", ff.RolloutPercentage())
	}
	if ff.IsActive() {
		t.Fatal("expected isActive false by default")
	}
	if ff.ID().String() == "" {
		t.Fatal("expected non-empty ID")
	}
}

func TestFeatureFlag_Evaluate_InactiveReturnsDefault(t *testing.T) {
	t.Parallel()

	ff, _ := domain.NewFeatureFlag("test", "test_flag", "desc", "bool", "false", 100)
	result := ff.Evaluate(map[string]string{"user_id": "user1"})
	if result != "false" {
		t.Fatalf("expected 'false', got %s", result)
	}
}

func TestFeatureFlag_Evaluate_RuleGroupMatch(t *testing.T) {
	t.Parallel()

	ff, _ := domain.NewFeatureFlag("test", "test_flag", "desc", "bool", "false", 0)
	ff.Activate()

	rg := domain.NewRuleGroup(ff.ID(), "beta", "true", 1)
	rg.AddCondition(domain.NewCondition("country", "eq", "US"))
	ff.AddRuleGroup(rg)

	result := ff.Evaluate(map[string]string{"country": "US"})
	if result != "true" {
		t.Fatalf("expected 'true', got %s", result)
	}
}

func TestFeatureFlag_Evaluate_RuleGroupNoMatch_FallsToDefault(t *testing.T) {
	t.Parallel()

	ff, _ := domain.NewFeatureFlag("test", "test_flag", "desc", "bool", "false", 0)
	ff.Activate()

	rg := domain.NewRuleGroup(ff.ID(), "beta", "true", 1)
	rg.AddCondition(domain.NewCondition("country", "eq", "US"))
	ff.AddRuleGroup(rg)

	result := ff.Evaluate(map[string]string{"country": "UK"})
	if result != "false" {
		t.Fatalf("expected 'false', got %s", result)
	}
}

func TestFeatureFlag_Evaluate_PriorityOrder(t *testing.T) {
	t.Parallel()

	ff, _ := domain.NewFeatureFlag("test", "test_flag", "desc", "string", "default", 0)
	ff.Activate()

	rg1 := domain.NewRuleGroup(ff.ID(), "low-priority", "variation-B", 2)
	rg1.AddCondition(domain.NewCondition("country", "eq", "US"))
	ff.AddRuleGroup(rg1)

	rg2 := domain.NewRuleGroup(ff.ID(), "high-priority", "variation-A", 1)
	rg2.AddCondition(domain.NewCondition("country", "eq", "US"))
	ff.AddRuleGroup(rg2)

	result := ff.Evaluate(map[string]string{"country": "US"})
	if result != "variation-A" {
		t.Fatalf("expected 'variation-A', got %s", result)
	}
}

func TestFeatureFlag_Evaluate_RolloutPercentage(t *testing.T) {
	t.Parallel()

	ff, _ := domain.NewFeatureFlag("test", "test_flag", "desc", "bool", "false", 100)
	ff.Activate()

	result := ff.Evaluate(map[string]string{"user_id": "user1"})
	if result != "true" {
		t.Fatalf("expected 'true' with 100%% rollout, got %s", result)
	}
}

func TestFeatureFlag_Evaluate_RolloutZero(t *testing.T) {
	t.Parallel()

	ff, _ := domain.NewFeatureFlag("test", "test_flag", "desc", "bool", "false", 0)
	ff.Activate()

	result := ff.Evaluate(map[string]string{"user_id": "user1"})
	if result != "false" {
		t.Fatalf("expected 'false' with 0%% rollout, got %s", result)
	}
}

func TestFeatureFlag_Toggle(t *testing.T) {
	t.Parallel()

	ff, _ := domain.NewFeatureFlag("test", "test_flag", "desc", "bool", "false", 100)

	ff.Activate()
	if !ff.IsActive() {
		t.Fatal("expected isActive true after Activate")
	}
	if len(ff.Events()) != 1 {
		t.Fatalf("expected 1 event, got %d", len(ff.Events()))
	}
	if ff.Events()[0].EventName() != "featureflag.toggled" {
		t.Fatalf("expected event featureflag.toggled, got %s", ff.Events()[0].EventName())
	}

	ff.Deactivate()
	if ff.IsActive() {
		t.Fatal("expected isActive false after Deactivate")
	}
	if len(ff.Events()) != 2 {
		t.Fatalf("expected 2 events, got %d", len(ff.Events()))
	}
}

func TestFeatureFlag_Evaluate_MultipleRuleGroups_AND(t *testing.T) {
	t.Parallel()

	ff, _ := domain.NewFeatureFlag("test", "test_flag", "desc", "bool", "false", 0)
	ff.Activate()

	rg := domain.NewRuleGroup(ff.ID(), "multi-condition", "true", 1)
	rg.AddCondition(domain.NewCondition("country", "eq", "US"))
	rg.AddCondition(domain.NewCondition("age", "gte", "18"))
	ff.AddRuleGroup(rg)

	// Both conditions match
	result := ff.Evaluate(map[string]string{"country": "US", "age": "25"})
	if result != "true" {
		t.Fatalf("expected 'true', got %s", result)
	}

	// One condition fails
	result = ff.Evaluate(map[string]string{"country": "US", "age": "16"})
	if result != "false" {
		t.Fatalf("expected 'false', got %s", result)
	}
}

func TestFeatureFlag_ReconstructWithRuleGroups(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	now := time.Now()

	rg := domain.ReconstructRuleGroup(uuid.New(), id, "beta", "true", 1, now, now, []domain.Condition{
		domain.ReconstructCondition(uuid.New(), uuid.New(), "country", "eq", "US"),
	})

	ff := domain.ReconstructFeatureFlag(id, now, now, nil, "test", "test_flag", "desc", "bool", "false", 50, true, []*domain.RuleGroup{rg})

	if ff.ID() != id {
		t.Fatal("expected matching ID")
	}
	if ff.Key() != "test_flag" {
		t.Fatalf("expected key test_flag, got %s", ff.Key())
	}
	if !ff.IsActive() {
		t.Fatal("expected isActive true")
	}
	if len(ff.RuleGroups()) != 1 {
		t.Fatalf("expected 1 rule group, got %d", len(ff.RuleGroups()))
	}
}
