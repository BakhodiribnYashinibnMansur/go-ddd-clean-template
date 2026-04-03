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

func TestIntegration_EvaluateFlag(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Create a flag with a rule group: region=eu -> "eu_variant".
	err := bc.CreateFlag.Handle(ctx, command.CreateCommand{
		Name:         "regional_feature",
		Key:          "regional_feature",
		Description:  "Region-based feature",
		FlagType:     "string",
		DefaultValue: "default_variant",
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
		Name:      "eu_users",
		Variation: "eu_variant",
		Priority:  1,
		Conditions: []command.ConditionInput{
			{Attribute: "region", Operator: "eq", Value: "eu"},
		},
	})
	if err != nil {
		t.Fatalf("CreateRuleGroup: %v", err)
	}

	bc.Evaluator.Invalidate(ctx)

	// Matching user attributes should return the rule group variation.
	result, err := bc.EvaluateFlag.Handle(ctx, query.EvaluateQuery{
		Key:       "regional_feature",
		UserAttrs: map[string]string{"region": "eu"},
	})
	if err != nil {
		t.Fatalf("EvaluateFlag: %v", err)
	}
	if result.Value != "eu_variant" {
		t.Errorf("expected value eu_variant, got %s", result.Value)
	}
	if result.FlagType != "string" {
		t.Errorf("expected flag_type string, got %s", result.FlagType)
	}
	if result.Key != "regional_feature" {
		t.Errorf("expected key regional_feature, got %s", result.Key)
	}

	// Non-matching attributes should return the default value.
	result2, err := bc.EvaluateFlag.Handle(ctx, query.EvaluateQuery{
		Key:       "regional_feature",
		UserAttrs: map[string]string{"region": "us"},
	})
	if err != nil {
		t.Fatalf("EvaluateFlag (non-matching): %v", err)
	}
	if result2.Value != "default_variant" {
		t.Errorf("expected default_variant for non-matching user, got %s", result2.Value)
	}
}

func TestIntegration_BatchEvaluateFlags(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Create two flags.
	for _, fc := range []command.CreateCommand{
		{
			Name:         "flag_alpha",
			Key:          "flag_alpha",
			Description:  "Alpha flag",
			FlagType:     "bool",
			DefaultValue: "false",
			IsActive:     true,
		},
		{
			Name:         "flag_beta",
			Key:          "flag_beta",
			Description:  "Beta flag",
			FlagType:     "string",
			DefaultValue: "off",
			IsActive:     true,
		},
	} {
		if err := bc.CreateFlag.Handle(ctx, fc); err != nil {
			t.Fatalf("CreateFlag %s: %v", fc.Key, err)
		}
	}

	bc.Evaluator.Invalidate(ctx)

	// Batch evaluate: two existing flags + one non-existent.
	batchResult, err := bc.BatchEvaluateFlag.Handle(ctx, query.BatchEvaluateQuery{
		Keys:      []string{"flag_alpha", "flag_beta", "nonexistent_flag"},
		UserAttrs: map[string]string{},
	})
	if err != nil {
		t.Fatalf("BatchEvaluateFlag: %v", err)
	}

	// Both existing flags should be present.
	if len(batchResult.Flags) != 2 {
		t.Fatalf("expected 2 flags in batch result, got %d", len(batchResult.Flags))
	}

	alpha, ok := batchResult.Flags["flag_alpha"]
	if !ok {
		t.Fatal("flag_alpha missing from batch result")
	}
	if alpha.Value != "false" {
		t.Errorf("expected flag_alpha value false, got %s", alpha.Value)
	}
	if alpha.FlagType != "bool" {
		t.Errorf("expected flag_alpha type bool, got %s", alpha.FlagType)
	}

	beta, ok := batchResult.Flags["flag_beta"]
	if !ok {
		t.Fatal("flag_beta missing from batch result")
	}
	if beta.Value != "off" {
		t.Errorf("expected flag_beta value off, got %s", beta.Value)
	}

	// Non-existent flag should be omitted.
	if _, ok := batchResult.Flags["nonexistent_flag"]; ok {
		t.Error("nonexistent_flag should not be in batch result")
	}
}

func TestIntegration_UpdateRuleGroup(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Create a flag with a rule group.
	err := bc.CreateFlag.Handle(ctx, command.CreateCommand{
		Name:         "updatable_flag",
		Key:          "updatable_flag",
		Description:  "Flag for rule group update test",
		FlagType:     "string",
		DefaultValue: "default",
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
		Name:      "original_name",
		Variation: "original_variation",
		Priority:  1,
		Conditions: []command.ConditionInput{
			{Attribute: "env", Operator: "eq", Value: "staging"},
		},
	})
	if err != nil {
		t.Fatalf("CreateRuleGroup: %v", err)
	}

	// Get the rule group ID.
	view, err := bc.GetFlag.Handle(ctx, query.GetQuery{ID: flagID})
	if err != nil {
		t.Fatalf("GetFlag: %v", err)
	}
	if len(view.RuleGroups) != 1 {
		t.Fatalf("expected 1 rule group, got %d", len(view.RuleGroups))
	}
	rgID := view.RuleGroups[0].ID

	// Update rule group name and variation.
	newName := "updated_name"
	newVariation := "updated_variation"
	err = bc.UpdateRuleGroup.Handle(ctx, command.UpdateRuleGroupCommand{
		ID:        rgID,
		Name:      &newName,
		Variation: &newVariation,
	})
	if err != nil {
		t.Fatalf("UpdateRuleGroup: %v", err)
	}

	// Verify via GetFlag.
	view2, err := bc.GetFlag.Handle(ctx, query.GetQuery{ID: flagID})
	if err != nil {
		t.Fatalf("GetFlag after update: %v", err)
	}
	if len(view2.RuleGroups) != 1 {
		t.Fatalf("expected 1 rule group after update, got %d", len(view2.RuleGroups))
	}
	rg := view2.RuleGroups[0]
	if rg.Name != "updated_name" {
		t.Errorf("expected rule group name updated_name, got %s", rg.Name)
	}
	if rg.Variation != "updated_variation" {
		t.Errorf("expected rule group variation updated_variation, got %s", rg.Variation)
	}
	// Conditions should remain unchanged.
	if len(rg.Conditions) != 1 {
		t.Fatalf("expected 1 condition preserved, got %d", len(rg.Conditions))
	}
}

func TestIntegration_RolloutPercentage(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Create an active bool flag with 100% rollout.
	err := bc.CreateFlag.Handle(ctx, command.CreateCommand{
		Name:              "full_rollout",
		Key:               "full_rollout",
		Description:       "Flag with 100% rollout",
		FlagType:          "bool",
		DefaultValue:      "true",
		IsActive:          true,
		RolloutPercentage: 100,
	})
	if err != nil {
		t.Fatalf("CreateFlag: %v", err)
	}

	bc.Evaluator.Invalidate(ctx)

	// Evaluate with a user_id; 100% rollout should always return "true".
	result, err := bc.EvaluateFlag.Handle(ctx, query.EvaluateQuery{
		Key:       "full_rollout",
		UserAttrs: map[string]string{"user_id": "user_123"},
	})
	if err != nil {
		t.Fatalf("EvaluateFlag: %v", err)
	}
	if result.Value != "true" {
		t.Errorf("expected value true for 100%% rollout, got %s", result.Value)
	}
	if result.FlagType != "bool" {
		t.Errorf("expected flag_type bool, got %s", result.FlagType)
	}
}

func TestIntegration_EvaluateInactiveFlag(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	// Create an inactive flag.
	err := bc.CreateFlag.Handle(ctx, command.CreateCommand{
		Name:         "inactive_flag",
		Key:          "inactive_flag",
		Description:  "This flag is not active",
		FlagType:     "bool",
		DefaultValue: "false",
		IsActive:     false,
	})
	if err != nil {
		t.Fatalf("CreateFlag: %v", err)
	}

	bc.Evaluator.Invalidate(ctx)

	// Evaluate inactive flag should return default value.
	result, err := bc.EvaluateFlag.Handle(ctx, query.EvaluateQuery{
		Key:       "inactive_flag",
		UserAttrs: map[string]string{"user_id": "user_456"},
	})
	if err != nil {
		t.Fatalf("EvaluateFlag: %v", err)
	}
	if result.Value != "false" {
		t.Errorf("expected default value false for inactive flag, got %s", result.Value)
	}
}
