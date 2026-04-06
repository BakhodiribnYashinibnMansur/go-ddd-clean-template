package command

import (
	"context"
	"errors"
	"testing"
	"time"

	ffentity "gct/internal/context/admin/generic/featureflag/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUpdateRuleGroupHandler_Handle(t *testing.T) {
	t.Parallel()

	rgID := ffentity.NewRuleGroupID()
	flagID := ffentity.NewFeatureFlagID()
	rg := ffentity.ReconstructRuleGroup(rgID.UUID(), flagID.UUID(), "old-name", "false", 1, time.Now(), time.Now(), nil)

	rgRepo := &mockRuleGroupRepo{
		findFn: func(_ context.Context, id ffentity.RuleGroupID) (*ffentity.RuleGroup, error) {
			if id == rgID {
				return rg, nil
			}
			return nil, ffentity.ErrRuleGroupNotFound
		},
	}
	eb := &mockEventBus{}
	handler := NewUpdateRuleGroupHandler(rgRepo, eb, &mockLogger{})

	newName := "new-name"
	newVariation := "true"
	newPriority := 5
	cmd := UpdateRuleGroupCommand{
		ID:        ffentity.RuleGroupID(rgID),
		Name:      &newName,
		Variation: &newVariation,
		Priority:  &newPriority,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if rgRepo.updated == nil {
		t.Fatal("expected rule group to be updated")
	}
	if rgRepo.updated.Name() != "new-name" {
		t.Errorf("expected name new-name, got %s", rgRepo.updated.Name())
	}
	if rgRepo.updated.Variation() != "true" {
		t.Errorf("expected variation true, got %s", rgRepo.updated.Variation())
	}
	if rgRepo.updated.Priority() != 5 {
		t.Errorf("expected priority 5, got %d", rgRepo.updated.Priority())
	}
	if len(eb.published) == 0 {
		t.Error("expected events to be published")
	}
}

func TestUpdateRuleGroupHandler_Handle_WithConditions(t *testing.T) {
	t.Parallel()

	rgID := ffentity.NewRuleGroupID()
	flagID := ffentity.NewFeatureFlagID()
	rg := ffentity.ReconstructRuleGroup(rgID.UUID(), flagID.UUID(), "rg", "false", 1, time.Now(), time.Now(), nil)

	rgRepo := &mockRuleGroupRepo{
		findFn: func(_ context.Context, _ ffentity.RuleGroupID) (*ffentity.RuleGroup, error) {
			return rg, nil
		},
	}
	handler := NewUpdateRuleGroupHandler(rgRepo, &mockEventBus{}, &mockLogger{})

	conditions := []ConditionInput{
		{Attribute: "plan", Operator: "eq", Value: "premium"},
		{Attribute: "region", Operator: "in", Value: "us,eu"},
	}
	cmd := UpdateRuleGroupCommand{
		ID:         ffentity.RuleGroupID(rgID),
		Conditions: &conditions,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if rgRepo.updated == nil {
		t.Fatal("expected rule group to be updated")
	}
	if len(rgRepo.updated.Conditions()) != 2 {
		t.Fatalf("expected 2 conditions, got %d", len(rgRepo.updated.Conditions()))
	}
	if rgRepo.updated.Conditions()[0].Attribute() != "plan" {
		t.Errorf("expected attribute plan, got %s", rgRepo.updated.Conditions()[0].Attribute())
	}
}

func TestUpdateRuleGroupHandler_Handle_InvalidOperator(t *testing.T) {
	t.Parallel()

	rgID := ffentity.NewRuleGroupID()
	flagID := ffentity.NewFeatureFlagID()
	rg := ffentity.ReconstructRuleGroup(rgID.UUID(), flagID.UUID(), "rg", "false", 1, time.Now(), time.Now(), nil)

	rgRepo := &mockRuleGroupRepo{
		findFn: func(_ context.Context, _ ffentity.RuleGroupID) (*ffentity.RuleGroup, error) {
			return rg, nil
		},
	}
	handler := NewUpdateRuleGroupHandler(rgRepo, &mockEventBus{}, &mockLogger{})

	conditions := []ConditionInput{
		{Attribute: "plan", Operator: "bad_op", Value: "premium"},
	}
	err := handler.Handle(context.Background(), UpdateRuleGroupCommand{
		ID:         ffentity.RuleGroupID(rgID),
		Conditions: &conditions,
	})
	if err == nil {
		t.Fatal("expected error for invalid operator, got nil")
	}
	if !errors.Is(err, ffentity.ErrInvalidOperator) {
		t.Fatalf("expected ErrInvalidOperator, got: %v", err)
	}
}

func TestUpdateRuleGroupHandler_Handle_NotFound(t *testing.T) {
	t.Parallel()

	rgRepo := &mockRuleGroupRepo{} // default returns ErrRuleGroupNotFound
	handler := NewUpdateRuleGroupHandler(rgRepo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), UpdateRuleGroupCommand{ID: ffentity.RuleGroupID(uuid.New())})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ffentity.ErrRuleGroupNotFound) {
		t.Fatalf("expected ErrRuleGroupNotFound, got: %v", err)
	}
}

func TestUpdateRuleGroupHandler_Handle_UpdateRepoError(t *testing.T) {
	t.Parallel()

	rgID := ffentity.NewRuleGroupID()
	flagID := ffentity.NewFeatureFlagID()
	rg := ffentity.ReconstructRuleGroup(rgID.UUID(), flagID.UUID(), "rg", "false", 1, time.Now(), time.Now(), nil)

	repoErr := errors.New("update failed")
	rgRepo := &mockRuleGroupRepo{
		findFn: func(_ context.Context, _ ffentity.RuleGroupID) (*ffentity.RuleGroup, error) {
			return rg, nil
		},
		updateFn: func(_ context.Context, _ *ffentity.RuleGroup) error {
			return repoErr
		},
	}
	handler := NewUpdateRuleGroupHandler(rgRepo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), UpdateRuleGroupCommand{ID: ffentity.RuleGroupID(rgID)})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got: %v", err)
	}
}
