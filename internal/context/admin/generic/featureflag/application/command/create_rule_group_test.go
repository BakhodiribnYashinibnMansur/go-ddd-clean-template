package command

import (
	"context"
	"errors"
	"testing"

	ffentity "gct/internal/context/admin/generic/featureflag/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateRuleGroupHandler_Handle(t *testing.T) {
	t.Parallel()

	flagID := ffentity.NewFeatureFlagID()
	flagRepo := &mockFeatureFlagRepo{
		findFn: func(_ context.Context, id ffentity.FeatureFlagID) (*ffentity.FeatureFlag, error) {
			if id == flagID {
				return newReconstructedFlag(flagID), nil
			}
			return nil, ffentity.ErrFeatureFlagNotFound
		},
	}
	rgRepo := &mockRuleGroupRepo{}
	eb := &mockEventBus{}
	handler := NewCreateRuleGroupHandler(flagRepo, rgRepo, eb, &mockLogger{})

	cmd := CreateRuleGroupCommand{
		FlagID:    ffentity.FeatureFlagID(flagID),
		Name:      "beta-users",
		Variation: "true",
		Priority:  1,
		Conditions: []ConditionInput{
			{Attribute: "plan", Operator: "eq", Value: "premium"},
		},
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if rgRepo.saved == nil {
		t.Fatal("expected rule group to be saved")
	}
	if rgRepo.saved.Name() != "beta-users" {
		t.Errorf("expected name beta-users, got %s", rgRepo.saved.Name())
	}
	if rgRepo.saved.Variation() != "true" {
		t.Errorf("expected variation true, got %s", rgRepo.saved.Variation())
	}
	if rgRepo.saved.Priority() != 1 {
		t.Errorf("expected priority 1, got %d", rgRepo.saved.Priority())
	}
	if len(rgRepo.saved.Conditions()) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(rgRepo.saved.Conditions()))
	}
	if rgRepo.saved.Conditions()[0].Attribute() != "plan" {
		t.Errorf("expected attribute plan, got %s", rgRepo.saved.Conditions()[0].Attribute())
	}
	if len(eb.published) == 0 {
		t.Error("expected events to be published")
	}
}

func TestCreateRuleGroupHandler_Handle_FlagNotFound(t *testing.T) {
	t.Parallel()

	flagRepo := &mockFeatureFlagRepo{} // default returns ErrFeatureFlagNotFound
	rgRepo := &mockRuleGroupRepo{}
	handler := NewCreateRuleGroupHandler(flagRepo, rgRepo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), CreateRuleGroupCommand{
		FlagID: ffentity.FeatureFlagID(uuid.New()),
		Name:   "test",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ffentity.ErrFeatureFlagNotFound) {
		t.Fatalf("expected ErrFeatureFlagNotFound, got: %v", err)
	}
	if rgRepo.saved != nil {
		t.Error("expected rule group NOT to be saved")
	}
}

func TestCreateRuleGroupHandler_Handle_InvalidOperator(t *testing.T) {
	t.Parallel()

	flagID := ffentity.NewFeatureFlagID()
	flagRepo := &mockFeatureFlagRepo{
		findFn: func(_ context.Context, id ffentity.FeatureFlagID) (*ffentity.FeatureFlag, error) {
			return newReconstructedFlag(id), nil
		},
	}
	rgRepo := &mockRuleGroupRepo{}
	handler := NewCreateRuleGroupHandler(flagRepo, rgRepo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), CreateRuleGroupCommand{
		FlagID:    ffentity.FeatureFlagID(flagID),
		Name:      "test",
		Variation: "true",
		Conditions: []ConditionInput{
			{Attribute: "plan", Operator: "invalid_op", Value: "premium"},
		},
	})
	if err == nil {
		t.Fatal("expected error for invalid operator, got nil")
	}
	if !errors.Is(err, ffentity.ErrInvalidOperator) {
		t.Fatalf("expected ErrInvalidOperator, got: %v", err)
	}
	if rgRepo.saved != nil {
		t.Error("expected rule group NOT to be saved")
	}
}

func TestCreateRuleGroupHandler_Handle_RepoError(t *testing.T) {
	t.Parallel()

	flagID := ffentity.NewFeatureFlagID()
	flagRepo := &mockFeatureFlagRepo{
		findFn: func(_ context.Context, _ ffentity.FeatureFlagID) (*ffentity.FeatureFlag, error) {
			return newReconstructedFlag(flagID), nil
		},
	}
	repoErr := errors.New("save failed")
	rgRepo := &mockRuleGroupRepo{
		saveFn: func(_ context.Context, _ *ffentity.RuleGroup) error {
			return repoErr
		},
	}
	handler := NewCreateRuleGroupHandler(flagRepo, rgRepo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), CreateRuleGroupCommand{
		FlagID:    ffentity.FeatureFlagID(flagID),
		Name:      "test",
		Variation: "true",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got: %v", err)
	}
}

func TestCreateRuleGroupHandler_Handle_MultipleConditions(t *testing.T) {
	t.Parallel()

	flagID := ffentity.NewFeatureFlagID()
	flagRepo := &mockFeatureFlagRepo{
		findFn: func(_ context.Context, _ ffentity.FeatureFlagID) (*ffentity.FeatureFlag, error) {
			return newReconstructedFlag(flagID), nil
		},
	}
	rgRepo := &mockRuleGroupRepo{}
	handler := NewCreateRuleGroupHandler(flagRepo, rgRepo, &mockEventBus{}, &mockLogger{})

	cmd := CreateRuleGroupCommand{
		FlagID:    ffentity.FeatureFlagID(flagID),
		Name:      "complex-rule",
		Variation: "variant-a",
		Priority:  5,
		Conditions: []ConditionInput{
			{Attribute: "plan", Operator: "eq", Value: "premium"},
			{Attribute: "age", Operator: "gte", Value: "18"},
			{Attribute: "country", Operator: "in", Value: "us,uk,ca"},
		},
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)
	if len(rgRepo.saved.Conditions()) != 3 {
		t.Fatalf("expected 3 conditions, got %d", len(rgRepo.saved.Conditions()))
	}
}
