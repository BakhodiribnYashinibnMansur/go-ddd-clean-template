package command

import (
	"context"
	"errors"
	"testing"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/application"
	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// --- Mocks ---

type mockRepo struct {
	saved   *domain.FeatureFlag
	updated *domain.FeatureFlag
	findFn  func(ctx context.Context, id uuid.UUID) (*domain.FeatureFlag, error)
}

func (m *mockRepo) Save(_ context.Context, e *domain.FeatureFlag) error {
	m.saved = e
	return nil
}

func (m *mockRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.FeatureFlag, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrFeatureFlagNotFound
}

func (m *mockRepo) Update(_ context.Context, e *domain.FeatureFlag) error {
	m.updated = e
	return nil
}

func (m *mockRepo) Delete(_ context.Context, _ uuid.UUID) error {
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

func (m *mockLogger) Debug(args ...any)                                          {}
func (m *mockLogger) Debugf(template string, args ...any)                        {}
func (m *mockLogger) Debugw(msg string, keysAndValues ...any)                    {}
func (m *mockLogger) Info(args ...any)                                           {}
func (m *mockLogger) Infof(template string, args ...any)                         {}
func (m *mockLogger) Infow(msg string, keysAndValues ...any)                     {}
func (m *mockLogger) Warn(args ...any)                                           {}
func (m *mockLogger) Warnf(template string, args ...any)                         {}
func (m *mockLogger) Warnw(msg string, keysAndValues ...any)                     {}
func (m *mockLogger) Error(args ...any)                                          {}
func (m *mockLogger) Errorf(template string, args ...any)                        {}
func (m *mockLogger) Errorw(msg string, keysAndValues ...any)                    {}
func (m *mockLogger) Fatal(args ...any)                                          {}
func (m *mockLogger) Fatalf(template string, args ...any)                        {}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any)                    {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any)               {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)                {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)                {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any)               {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any)               {}

// --- Tests ---

func TestCreateHandler_Handle(t *testing.T) {
	repo := &mockRepo{}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateHandler(repo, eb, log)

	cmd := CreateCommand{
		Name:              "dark_mode",
		Description:       "Enable dark mode for users",
		Enabled:           true,
		RolloutPercentage: 50,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.saved == nil {
		t.Fatal("expected feature flag to be saved")
	}
	if repo.saved.Name() != "dark_mode" {
		t.Errorf("expected name dark_mode, got %s", repo.saved.Name())
	}
	if repo.saved.Description() != "Enable dark mode for users" {
		t.Errorf("expected description, got %s", repo.saved.Description())
	}
	if repo.saved.Enabled() != true {
		t.Errorf("expected enabled true, got %v", repo.saved.Enabled())
	}
	if repo.saved.RolloutPercentage() != 50 {
		t.Errorf("expected rollout 50, got %d", repo.saved.RolloutPercentage())
	}
}

func TestCreateHandler_DisabledFlag(t *testing.T) {
	repo := &mockRepo{}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateHandler(repo, eb, log)

	err := handler.Handle(context.Background(), CreateCommand{
		Name:              "new_feature",
		Description:       "Coming soon",
		Enabled:           false,
		RolloutPercentage: 0,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.saved == nil {
		t.Fatal("expected feature flag to be saved")
	}
	if repo.saved.Enabled() {
		t.Error("expected flag to be disabled")
	}
	if repo.saved.RolloutPercentage() != 0 {
		t.Errorf("expected rollout 0, got %d", repo.saved.RolloutPercentage())
	}
}

func TestCreateHandler_RepoError(t *testing.T) {
	repoErr := errors.New("repo save failed")
	errR := &errorRepo{saveErr: repoErr}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateHandler(errR, eb, log)
	err := handler.Handle(context.Background(), CreateCommand{
		Name: "f", Description: "d", Enabled: true, RolloutPercentage: 100,
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo save error, got: %v", err)
	}
}
