package featureflag

import (
	"context"
	"testing"

	"gct/internal/featureflag"
	"gct/internal/featureflag/application/command"
	"gct/internal/featureflag/application/query"
	"gct/internal/featureflag/domain"
	"gct/internal/shared/infrastructure/eventbus"
	"gct/internal/shared/infrastructure/logger"
	"gct/test/integration/common/setup"
)

func newTestBC(t *testing.T) *featureflag.BoundedContext {
	t.Helper()
	eb := eventbus.NewInMemoryEventBus()
	l := logger.New("error")
	bc, err := featureflag.NewBoundedContext(context.Background(), setup.TestPG.Pool, eb, l)
	if err != nil {
		t.Fatalf("NewBoundedContext: %v", err)
	}
	return bc
}

func TestIntegration_CreateAndGetFeatureFlag(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateFlag.Handle(ctx, command.CreateCommand{
		Name:              "dark_mode",
		Key:               "dark_mode",
		Description:       "Enable dark mode UI",
		FlagType:          "bool",
		DefaultValue:      "false",
		IsActive:          true,
		RolloutPercentage: 50,
	})
	if err != nil {
		t.Fatalf("CreateFlag: %v", err)
	}

	result, err := bc.ListFlags.Handle(ctx, query.ListQuery{
		Filter: domain.FeatureFlagFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListFlags: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 feature flag, got %d", result.Total)
	}

	f := result.Flags[0]
	if f.Name != "dark_mode" {
		t.Errorf("expected name dark_mode, got %s", f.Name)
	}
	if f.Key != "dark_mode" {
		t.Errorf("expected key dark_mode, got %s", f.Key)
	}
	if f.FlagType != "bool" {
		t.Errorf("expected flag_type bool, got %s", f.FlagType)
	}
	if f.DefaultValue != "false" {
		t.Errorf("expected default_value false, got %s", f.DefaultValue)
	}
	if !f.IsActive {
		t.Error("expected flag to be active")
	}

	view, err := bc.GetFlag.Handle(ctx, query.GetQuery{ID: f.ID})
	if err != nil {
		t.Fatalf("GetFlag: %v", err)
	}
	if view.ID != f.ID {
		t.Errorf("ID mismatch: %s vs %s", view.ID, f.ID)
	}
}

func TestIntegration_UpdateFeatureFlag(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateFlag.Handle(ctx, command.CreateCommand{
		Name:              "new_checkout",
		Key:               "new_checkout",
		Description:       "New checkout flow",
		FlagType:          "bool",
		DefaultValue:      "false",
		IsActive:          false,
		RolloutPercentage: 0,
	})
	if err != nil {
		t.Fatalf("CreateFlag: %v", err)
	}

	list, err := bc.ListFlags.Handle(ctx, query.ListQuery{
		Filter: domain.FeatureFlagFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListFlags: %v", err)
	}
	fID := list.Flags[0].ID

	newName := "updated_checkout"
	newActive := true
	newRollout := 75
	err = bc.UpdateFlag.Handle(ctx, command.UpdateCommand{
		ID:                fID,
		Name:              &newName,
		IsActive:          &newActive,
		RolloutPercentage: &newRollout,
	})
	if err != nil {
		t.Fatalf("UpdateFlag: %v", err)
	}

	view, err := bc.GetFlag.Handle(ctx, query.GetQuery{ID: fID})
	if err != nil {
		t.Fatalf("GetFlag after update: %v", err)
	}
	if view.Name != "updated_checkout" {
		t.Errorf("name not updated, got %s", view.Name)
	}
	if !view.IsActive {
		t.Error("flag should be active after update")
	}
}

func TestIntegration_DeleteFeatureFlag(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateFlag.Handle(ctx, command.CreateCommand{
		Name:              "to_delete",
		Key:               "to_delete",
		Description:       "Will be deleted",
		FlagType:          "bool",
		DefaultValue:      "false",
		IsActive:          true,
		RolloutPercentage: 100,
	})
	if err != nil {
		t.Fatalf("CreateFlag: %v", err)
	}

	list, err := bc.ListFlags.Handle(ctx, query.ListQuery{
		Filter: domain.FeatureFlagFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListFlags: %v", err)
	}
	fID := list.Flags[0].ID

	err = bc.DeleteFlag.Handle(ctx, command.DeleteCommand{ID: fID})
	if err != nil {
		t.Fatalf("DeleteFlag: %v", err)
	}

	list2, err := bc.ListFlags.Handle(ctx, query.ListQuery{
		Filter: domain.FeatureFlagFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListFlags after delete: %v", err)
	}
	if list2.Total != 0 {
		t.Errorf("expected 0 feature flags after delete, got %d", list2.Total)
	}
}

func TestIntegration_RuleGroupCRUD(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Create a flag to attach rule groups to.
	err := bc.CreateFlag.Handle(ctx, command.CreateCommand{
		Name:         "premium_feature",
		Key:          "premium_feature",
		Description:  "Premium-only feature",
		FlagType:     "bool",
		DefaultValue: "false",
		IsActive:     true,
	})
	if err != nil {
		t.Fatalf("CreateFlag: %v", err)
	}

	list, err := bc.ListFlags.Handle(ctx, query.ListQuery{
		Filter: domain.FeatureFlagFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListFlags: %v", err)
	}
	flagID := list.Flags[0].ID

	// Create a rule group with 2 conditions.
	err = bc.CreateRuleGroup.Handle(ctx, command.CreateRuleGroupCommand{
		FlagID:    flagID,
		Name:      "premium_users",
		Variation: "true",
		Priority:  1,
		Conditions: []command.ConditionInput{
			{Attribute: "plan", Operator: "eq", Value: "premium"},
			{Attribute: "country", Operator: "eq", Value: "US"},
		},
	})
	if err != nil {
		t.Fatalf("CreateRuleGroup: %v", err)
	}

	// Verify via GetFlag that the rule group and conditions are returned.
	view, err := bc.GetFlag.Handle(ctx, query.GetQuery{ID: flagID})
	if err != nil {
		t.Fatalf("GetFlag: %v", err)
	}
	if len(view.RuleGroups) != 1 {
		t.Fatalf("expected 1 rule group, got %d", len(view.RuleGroups))
	}
	rg := view.RuleGroups[0]
	if rg.Name != "premium_users" {
		t.Errorf("expected rule group name premium_users, got %s", rg.Name)
	}
	if rg.Variation != "true" {
		t.Errorf("expected variation true, got %s", rg.Variation)
	}
	if len(rg.Conditions) != 2 {
		t.Fatalf("expected 2 conditions, got %d", len(rg.Conditions))
	}

	// Delete the rule group.
	err = bc.DeleteRuleGroup.Handle(ctx, command.DeleteRuleGroupCommand{ID: rg.ID})
	if err != nil {
		t.Fatalf("DeleteRuleGroup: %v", err)
	}

	// Verify the rule group is gone.
	view2, err := bc.GetFlag.Handle(ctx, query.GetQuery{ID: flagID})
	if err != nil {
		t.Fatalf("GetFlag after delete: %v", err)
	}
	if len(view2.RuleGroups) != 0 {
		t.Errorf("expected 0 rule groups after delete, got %d", len(view2.RuleGroups))
	}
}

func TestIntegration_Evaluator(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Create a flag with a rule group: plan=premium -> true.
	err := bc.CreateFlag.Handle(ctx, command.CreateCommand{
		Name:         "premium_gate",
		Key:          "premium_gate",
		Description:  "Gate for premium users",
		FlagType:     "bool",
		DefaultValue: "false",
		IsActive:     true,
	})
	if err != nil {
		t.Fatalf("CreateFlag: %v", err)
	}

	list, err := bc.ListFlags.Handle(ctx, query.ListQuery{
		Filter: domain.FeatureFlagFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListFlags: %v", err)
	}
	flagID := list.Flags[0].ID

	err = bc.CreateRuleGroup.Handle(ctx, command.CreateRuleGroupCommand{
		FlagID:    flagID,
		Name:      "premium_rule",
		Variation: "true",
		Priority:  1,
		Conditions: []command.ConditionInput{
			{Attribute: "plan", Operator: "eq", Value: "premium"},
		},
	})
	if err != nil {
		t.Fatalf("CreateRuleGroup: %v", err)
	}

	// Invalidate the evaluator cache so it picks up the new flag and rule group.
	bc.Evaluator.Invalidate(ctx)

	// Premium user should get true.
	if !bc.Evaluator.IsEnabled(ctx, "premium_gate", map[string]string{"plan": "premium"}) {
		t.Error("expected premium user to get true")
	}

	// Free user should get false (default value).
	if bc.Evaluator.IsEnabled(ctx, "premium_gate", map[string]string{"plan": "free"}) {
		t.Error("expected free user to get false")
	}

	// Nonexistent flag should return false.
	if bc.Evaluator.IsEnabled(ctx, "nonexistent_flag", map[string]string{"plan": "premium"}) {
		t.Error("expected nonexistent flag to return false")
	}
}
