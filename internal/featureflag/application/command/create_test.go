package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/application"
	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// --- Mocks ---

type mockFeatureFlagRepo struct {
	saved   *domain.FeatureFlag
	updated *domain.FeatureFlag
	deleted uuid.UUID
	findFn  func(ctx context.Context, id uuid.UUID) (*domain.FeatureFlag, error)
	saveFn  func(ctx context.Context, e *domain.FeatureFlag) error
	updateFn func(ctx context.Context, e *domain.FeatureFlag) error
	deleteFn func(ctx context.Context, id uuid.UUID) error
}

func (m *mockFeatureFlagRepo) Save(ctx context.Context, e *domain.FeatureFlag) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, e)
	}
	m.saved = e
	return nil
}

func (m *mockFeatureFlagRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.FeatureFlag, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrFeatureFlagNotFound
}

func (m *mockFeatureFlagRepo) FindByKey(_ context.Context, _ string) (*domain.FeatureFlag, error) {
	return nil, domain.ErrFeatureFlagNotFound
}

func (m *mockFeatureFlagRepo) Update(ctx context.Context, e *domain.FeatureFlag) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, e)
	}
	m.updated = e
	return nil
}

func (m *mockFeatureFlagRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	m.deleted = id
	return nil
}

func (m *mockFeatureFlagRepo) FindAll(_ context.Context) ([]*domain.FeatureFlag, error) {
	return nil, nil
}

type mockRuleGroupRepo struct {
	saved    *domain.RuleGroup
	updated  *domain.RuleGroup
	deleted  uuid.UUID
	findFn   func(ctx context.Context, id uuid.UUID) (*domain.RuleGroup, error)
	saveFn   func(ctx context.Context, rg *domain.RuleGroup) error
	updateFn func(ctx context.Context, rg *domain.RuleGroup) error
	deleteFn func(ctx context.Context, id uuid.UUID) error
}

func (m *mockRuleGroupRepo) Save(ctx context.Context, rg *domain.RuleGroup) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, rg)
	}
	m.saved = rg
	return nil
}

func (m *mockRuleGroupRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.RuleGroup, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrRuleGroupNotFound
}

func (m *mockRuleGroupRepo) Update(ctx context.Context, rg *domain.RuleGroup) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, rg)
	}
	m.updated = rg
	return nil
}

func (m *mockRuleGroupRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	m.deleted = id
	return nil
}

func (m *mockRuleGroupRepo) FindByFlagID(_ context.Context, _ uuid.UUID) ([]*domain.RuleGroup, error) {
	return nil, nil
}

func (m *mockRuleGroupRepo) SaveCondition(_ context.Context, _ uuid.UUID, _ domain.Condition) error {
	return nil
}

func (m *mockRuleGroupRepo) DeleteConditionsByRuleGroupID(_ context.Context, _ uuid.UUID) error {
	return nil
}

type mockEventBus struct {
	published []shared.DomainEvent
}

func (m *mockEventBus) Publish(_ context.Context, events ...shared.DomainEvent) error {
	m.published = append(m.published, events...)
	return nil
}

func (m *mockEventBus) Subscribe(_ string, _ application.EventHandler) error { return nil }

type mockLogger struct{}

func (m *mockLogger) Debug(_ ...any)                                {}
func (m *mockLogger) Debugf(_ string, _ ...any)                     {}
func (m *mockLogger) Debugw(_ string, _ ...any)                     {}
func (m *mockLogger) Info(_ ...any)                                 {}
func (m *mockLogger) Infof(_ string, _ ...any)                      {}
func (m *mockLogger) Infow(_ string, _ ...any)                      {}
func (m *mockLogger) Warn(_ ...any)                                 {}
func (m *mockLogger) Warnf(_ string, _ ...any)                      {}
func (m *mockLogger) Warnw(_ string, _ ...any)                      {}
func (m *mockLogger) Error(_ ...any)                                {}
func (m *mockLogger) Errorf(_ string, _ ...any)                     {}
func (m *mockLogger) Errorw(_ string, _ ...any)                     {}
func (m *mockLogger) Fatal(_ ...any)                                {}
func (m *mockLogger) Fatalf(_ string, _ ...any)                     {}
func (m *mockLogger) Fatalw(_ string, _ ...any)                     {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)   {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)   {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any)  {}

// helper to create a reconstructed feature flag for FindByID mocks
func newReconstructedFlag(id uuid.UUID) *domain.FeatureFlag {
	return domain.ReconstructFeatureFlag(
		id, time.Now(), time.Now(), nil,
		"test-flag", "test_key", "desc", "bool", "false", 50, true, nil,
	)
}

// --- Tests ---

func TestCreateHandler_Handle(t *testing.T) {
	repo := &mockFeatureFlagRepo{}
	eb := &mockEventBus{}
	handler := NewCreateHandler(repo, eb, &mockLogger{})

	cmd := CreateCommand{
		Name:              "dark-mode",
		Key:               "dark_mode",
		Description:       "Enable dark mode",
		FlagType:          "bool",
		DefaultValue:      "false",
		RolloutPercentage: 50,
		IsActive:          false,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.saved == nil {
		t.Fatal("expected feature flag to be saved")
	}
	if repo.saved.Name() != "dark-mode" {
		t.Errorf("expected name dark-mode, got %s", repo.saved.Name())
	}
	if repo.saved.Key() != "dark_mode" {
		t.Errorf("expected key dark_mode, got %s", repo.saved.Key())
	}
	if repo.saved.FlagType() != "bool" {
		t.Errorf("expected flagType bool, got %s", repo.saved.FlagType())
	}
	if repo.saved.DefaultValue() != "false" {
		t.Errorf("expected defaultValue false, got %s", repo.saved.DefaultValue())
	}
	if repo.saved.RolloutPercentage() != 50 {
		t.Errorf("expected rolloutPercentage 50, got %d", repo.saved.RolloutPercentage())
	}
	if repo.saved.IsActive() {
		t.Error("expected flag to be inactive")
	}
	if len(eb.published) == 0 {
		t.Error("expected events to be published")
	}
}

func TestCreateHandler_Handle_Active(t *testing.T) {
	repo := &mockFeatureFlagRepo{}
	eb := &mockEventBus{}
	handler := NewCreateHandler(repo, eb, &mockLogger{})

	cmd := CreateCommand{
		Name:     "feature-x",
		Key:      "feature_x",
		FlagType: "bool",
		IsActive: true,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.saved == nil {
		t.Fatal("expected feature flag to be saved")
	}
	if !repo.saved.IsActive() {
		t.Error("expected flag to be active")
	}
}

func TestCreateHandler_Handle_RepoError(t *testing.T) {
	repoErr := errors.New("db failure")
	repo := &mockFeatureFlagRepo{
		saveFn: func(_ context.Context, _ *domain.FeatureFlag) error {
			return repoErr
		},
	}
	handler := NewCreateHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), CreateCommand{
		Name:     "test",
		Key:      "test_key",
		FlagType: "bool",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got: %v", err)
	}
}
