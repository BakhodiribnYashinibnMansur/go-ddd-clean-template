package query

import (
	"context"
	"errors"
	"gct/internal/kernel/infrastructure/logger"
	"testing"
	"time"

	"gct/internal/context/content/generic/translation/domain"

	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockReadRepo struct {
	view  *domain.TranslationView
	views []*domain.TranslationView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id domain.TranslationID) (*domain.TranslationView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrTranslationNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ domain.TranslationFilter) ([]*domain.TranslationView, int64, error) {
	return m.views, m.total, nil
}

type errorReadRepo struct{ err error }

func (m *errorReadRepo) FindByID(_ context.Context, _ domain.TranslationID) (*domain.TranslationView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ domain.TranslationFilter) ([]*domain.TranslationView, int64, error) {
	return nil, 0, m.err
}

var errRepo = errors.New("repo failure")

// --- Tests: GetTranslation ---

func TestGetTranslationHandler_Handle(t *testing.T) {
	t.Parallel()

	id := domain.NewTranslationID()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &domain.TranslationView{
			ID:        id,
			Key:       "welcome",
			Language:  "en",
			Value:     "Welcome!",
			Group:     "auth",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := NewGetTranslationHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetTranslationQuery{ID: domain.TranslationID(id)})
	require.NoError(t, err)
	if result == nil {
		t.Fatal("expected result")
	}
	if result.Key != "welcome" {
		t.Errorf("expected key welcome, got %s", result.Key)
	}
	if result.Language != "en" {
		t.Errorf("expected language en, got %s", result.Language)
	}
	if result.Value != "Welcome!" {
		t.Errorf("expected value Welcome!, got %s", result.Value)
	}
	if result.Group != "auth" {
		t.Errorf("expected group auth, got %s", result.Group)
	}
}

func TestGetTranslationHandler_NotFound(t *testing.T) {
	t.Parallel()

	readRepo := &mockReadRepo{}
	handler := NewGetTranslationHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetTranslationQuery{ID: domain.NewTranslationID()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetTranslationHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepo}
	handler := NewGetTranslationHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetTranslationQuery{ID: domain.NewTranslationID()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestGetTranslationHandler_AllFieldsMapped(t *testing.T) {
	t.Parallel()

	id := domain.NewTranslationID()
	now := time.Now()

	readRepo := &mockReadRepo{
		view: &domain.TranslationView{
			ID:        id,
			Key:       "btn_submit",
			Language:  "de",
			Value:     "Absenden",
			Group:     "forms",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := NewGetTranslationHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetTranslationQuery{ID: domain.TranslationID(id)})
	require.NoError(t, err)
	if result.ID != id {
		t.Error("ID not mapped correctly")
	}
	if result.Key != "btn_submit" {
		t.Error("key not mapped")
	}
	if result.Language != "de" {
		t.Error("language not mapped")
	}
	if result.Value != "Absenden" {
		t.Error("value not mapped")
	}
	if result.Group != "forms" {
		t.Error("group not mapped")
	}
	if result.CreatedAt.IsZero() {
		t.Error("created_at not mapped")
	}
	if result.UpdatedAt.IsZero() {
		t.Error("updated_at not mapped")
	}
}
