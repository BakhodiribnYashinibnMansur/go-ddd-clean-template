package command

import (
	"context"
	"errors"
	"testing"

	"gct/internal/shared/application"
	shared "gct/internal/shared/domain"
	"gct/internal/translation/domain"

	"github.com/google/uuid"
)

// --- Mocks ---

type mockRepo struct {
	saved   *domain.Translation
	updated *domain.Translation
	findFn  func(ctx context.Context, id uuid.UUID) (*domain.Translation, error)
}

func (m *mockRepo) Save(_ context.Context, e *domain.Translation) error {
	m.saved = e
	return nil
}

func (m *mockRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Translation, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrTranslationNotFound
}

func (m *mockRepo) Update(_ context.Context, e *domain.Translation) error {
	m.updated = e
	return nil
}

func (m *mockRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockRepo) List(_ context.Context, _ domain.TranslationFilter) ([]*domain.Translation, int64, error) {
	return nil, 0, nil
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

func TestCreateTranslationHandler_Handle(t *testing.T) {
	repo := &mockRepo{}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateTranslationHandler(repo, eb, log)

	cmd := CreateTranslationCommand{
		Key:      "welcome_message",
		Language: "en",
		Value:    "Welcome!",
		Group:    "auth",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.saved == nil {
		t.Fatal("expected translation to be saved")
	}
	if repo.saved.Key() != "welcome_message" {
		t.Errorf("expected key welcome_message, got %s", repo.saved.Key())
	}
	if repo.saved.Language() != "en" {
		t.Errorf("expected language en, got %s", repo.saved.Language())
	}
	if repo.saved.Value() != "Welcome!" {
		t.Errorf("expected value Welcome!, got %s", repo.saved.Value())
	}
	if repo.saved.Group() != "auth" {
		t.Errorf("expected group auth, got %s", repo.saved.Group())
	}
}

func TestCreateTranslationHandler_MinimalFields(t *testing.T) {
	repo := &mockRepo{}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateTranslationHandler(repo, eb, log)

	err := handler.Handle(context.Background(), CreateTranslationCommand{
		Key:      "btn_ok",
		Language: "fr",
		Value:    "OK",
		Group:    "common",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.saved == nil {
		t.Fatal("expected translation to be saved")
	}
	if repo.saved.Key() != "btn_ok" {
		t.Errorf("expected key btn_ok, got %s", repo.saved.Key())
	}
}

func TestCreateTranslationHandler_RepoError(t *testing.T) {
	repoErr := errors.New("repo save failed")
	errRepo := &errorRepo{saveErr: repoErr}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateTranslationHandler(errRepo, eb, log)
	err := handler.Handle(context.Background(), CreateTranslationCommand{
		Key: "k", Language: "en", Value: "v", Group: "g",
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo save error, got: %v", err)
	}
}
