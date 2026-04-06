package query

import (
	"context"
	"errors"
	"gct/internal/kernel/infrastructure/logger"
	"testing"
	"time"

	settingentity "gct/internal/context/iam/generic/usersetting/domain/entity"
	settingrepo "gct/internal/context/iam/generic/usersetting/domain/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockReadRepo struct {
	view  *settingrepo.UserSettingView
	views []*settingrepo.UserSettingView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id settingentity.UserSettingID) (*settingrepo.UserSettingView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, settingentity.ErrUserSettingNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ settingrepo.UserSettingFilter) ([]*settingrepo.UserSettingView, int64, error) {
	return m.views, m.total, nil
}

type errorReadRepo struct{ err error }

func (m *errorReadRepo) FindByID(_ context.Context, _ settingentity.UserSettingID) (*settingrepo.UserSettingView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ settingrepo.UserSettingFilter) ([]*settingrepo.UserSettingView, int64, error) {
	return nil, 0, m.err
}

var errRepo = errors.New("repo failure")

// --- Tests: GetUserSetting ---

func TestGetUserSettingHandler_Handle(t *testing.T) {
	t.Parallel()

	id := settingentity.NewUserSettingID()
	userID := uuid.New()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &settingrepo.UserSettingView{
			ID:        id,
			UserID:    userID,
			Key:       "theme",
			Value:     "dark",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	handler := NewGetUserSettingHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetUserSettingQuery{ID: settingentity.UserSettingID(id)})
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
	_, err := handler.Handle(context.Background(), GetUserSettingQuery{ID: settingentity.NewUserSettingID()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetUserSettingHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepo}
	handler := NewGetUserSettingHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetUserSettingQuery{ID: settingentity.NewUserSettingID()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
