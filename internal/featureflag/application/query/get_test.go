package query

import (
	"gct/internal/shared/infrastructure/logger"
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/featureflag/domain"

	"github.com/google/uuid"
)

// --- Mocks ---

type mockReadRepo struct {
	findByIDFn func(ctx context.Context, id uuid.UUID) (*domain.FeatureFlagView, error)
	listFn     func(ctx context.Context, filter domain.FeatureFlagFilter) ([]*domain.FeatureFlagView, int64, error)
}

func (m *mockReadRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.FeatureFlagView, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, domain.ErrFeatureFlagNotFound
}

func (m *mockReadRepo) List(ctx context.Context, filter domain.FeatureFlagFilter) ([]*domain.FeatureFlagView, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return nil, 0, nil
}

// --- Tests ---

func TestGetHandler_Handle(t *testing.T) {
	flagID := uuid.New()
	now := time.Now().Format(time.RFC3339)

	readRepo := &mockReadRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.FeatureFlagView, error) {
			if id == flagID {
				return &domain.FeatureFlagView{
					ID:                flagID,
					Name:              "dark-mode",
					Key:               "dark_mode",
					Description:       "Enable dark mode",
					FlagType:          "bool",
					DefaultValue:      "false",
					RolloutPercentage: 50,
					IsActive:          true,
					RuleGroups: []domain.RuleGroupView{
						{
							ID:        uuid.New(),
							Name:      "beta",
							Variation: "true",
							Priority:  1,
							Conditions: []domain.ConditionView{
								{ID: uuid.New(), Attribute: "plan", Operator: "eq", Value: "premium"},
							},
							CreatedAt: now,
							UpdatedAt: now,
						},
					},
					CreatedAt: now,
					UpdatedAt: now,
				}, nil
			}
			return nil, domain.ErrFeatureFlagNotFound
		},
	}

	handler := NewGetHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetQuery{ID: flagID})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.ID != flagID {
		t.Errorf("expected ID %s, got %s", flagID, result.ID)
	}
	if result.Name != "dark-mode" {
		t.Errorf("expected name dark-mode, got %s", result.Name)
	}
	if result.Key != "dark_mode" {
		t.Errorf("expected key dark_mode, got %s", result.Key)
	}
	if !result.IsActive {
		t.Error("expected IsActive true")
	}
	if len(result.RuleGroups) != 1 {
		t.Fatalf("expected 1 rule group, got %d", len(result.RuleGroups))
	}
	if result.RuleGroups[0].Name != "beta" {
		t.Errorf("expected rule group name beta, got %s", result.RuleGroups[0].Name)
	}
	if len(result.RuleGroups[0].Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(result.RuleGroups[0].Conditions))
	}
	if result.RuleGroups[0].Conditions[0].Attribute != "plan" {
		t.Errorf("expected condition attribute plan, got %s", result.RuleGroups[0].Conditions[0].Attribute)
	}
}

func TestGetHandler_Handle_NotFound(t *testing.T) {
	readRepo := &mockReadRepo{} // default returns ErrFeatureFlagNotFound
	handler := NewGetHandler(readRepo, logger.Noop())

	result, err := handler.Handle(context.Background(), GetQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrFeatureFlagNotFound) {
		t.Fatalf("expected ErrFeatureFlagNotFound, got: %v", err)
	}
	if result != nil {
		t.Error("expected nil result")
	}
}

func TestGetHandler_Handle_RepoError(t *testing.T) {
	repoErr := errors.New("db connection failed")
	readRepo := &mockReadRepo{
		findByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.FeatureFlagView, error) {
			return nil, repoErr
		},
	}
	handler := NewGetHandler(readRepo, logger.Noop())

	result, err := handler.Handle(context.Background(), GetQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got: %v", err)
	}
	if result != nil {
		t.Error("expected nil result")
	}
}

func TestGetHandler_Handle_EmptyRuleGroups(t *testing.T) {
	flagID := uuid.New()
	now := time.Now().Format(time.RFC3339)

	readRepo := &mockReadRepo{
		findByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.FeatureFlagView, error) {
			return &domain.FeatureFlagView{
				ID:         flagID,
				Name:       "simple-flag",
				Key:        "simple_flag",
				FlagType:   "bool",
				RuleGroups: []domain.RuleGroupView{},
				CreatedAt:  now,
				UpdatedAt:  now,
			}, nil
		},
	}

	handler := NewGetHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetQuery{ID: flagID})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(result.RuleGroups) != 0 {
		t.Errorf("expected 0 rule groups, got %d", len(result.RuleGroups))
	}
}
