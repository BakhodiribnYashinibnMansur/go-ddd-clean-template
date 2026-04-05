package query

import (
	"context"
	"errors"
	"gct/internal/kernel/infrastructure/logger"
	"testing"
	"time"

	"gct/internal/context/iam/generic/usersetting/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockReadRepo struct {
	view  *domain.UserSettingView
	views []*domain.UserSettingView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id domain.UserSettingID) (*domain.UserSettingView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrUserSettingNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ domain.UserSettingFilter) ([]*domain.UserSettingView, int64, error) {
	return m.views, m.total, nil
}

type errorReadRepo struct{ err error }

func (m *errorReadRepo) FindByID(_ context.Context, _ domain.UserSettingID) (*domain.UserSettingView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ domain.UserSettingFilter) ([]*domain.UserSettingView, int64, error) {
	return nil, 0, m.err
}

var errRepo = errors.New("repo failure")

// --- Tests: GetUserSetting ---

func TestGetUserSettingHandler_Handle(t *testing.T) {
	t.Parallel()

	id := domain.NewUserSettingID()
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
	result, err := handler.Handle(context.Background(), GetUserSettingQuery{ID: domain.UserSettingID(id)})
	require.NoError(t, err)
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
	t.Parallel()

	readRepo := &mockReadRepo{}
	handler := NewGetUserSettingHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetUserSettingQuery{ID: domain.NewUserSettingID()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetUserSettingHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepo}
	handler := NewGetUserSettingHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetUserSettingQuery{ID: domain.NewUserSettingID()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
