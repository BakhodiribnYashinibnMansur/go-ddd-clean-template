package query

import (
	"gct/internal/shared/infrastructure/logger"
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/usersetting/domain"

	"github.com/google/uuid"
)

// --- Mocks ---

type mockReadRepo struct {
	view  *domain.UserSettingView
	views []*domain.UserSettingView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.UserSettingView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrUserSettingNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ domain.UserSettingFilter) ([]*domain.UserSettingView, int64, error) {
	return m.views, m.total, nil
}

type errorReadRepo struct{ err error }

func (m *errorReadRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.UserSettingView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ domain.UserSettingFilter) ([]*domain.UserSettingView, int64, error) {
	return nil, 0, m.err
}

var errRepo = errors.New("repo failure")

// --- Tests: GetUserSetting ---

func TestGetUserSettingHandler_Handle(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &domain.UserSettingView{
			ID:        id,
			UserID:    userID,
			Key:       "theme",
			Value:     "dark",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := NewGetUserSettingHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetUserSettingQuery{ID: id})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if result.Key != "theme" {
		t.Errorf("expected key 'theme', got %s", result.Key)
	}
	if result.Value != "dark" {
		t.Errorf("expected value 'dark', got %s", result.Value)
	}
	if result.UserID != userID {
		t.Errorf("expected userID %s, got %s", userID, result.UserID)
	}
}

func TestGetUserSettingHandler_NotFound(t *testing.T) {
	readRepo := &mockReadRepo{}
	handler := NewGetUserSettingHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetUserSettingQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetUserSettingHandler_RepoError(t *testing.T) {
	readRepo := &errorReadRepo{err: errRepo}
	handler := NewGetUserSettingHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetUserSettingQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
