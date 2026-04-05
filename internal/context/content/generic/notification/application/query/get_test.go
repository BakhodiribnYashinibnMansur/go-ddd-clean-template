package query

import (
	"gct/internal/kernel/infrastructure/logger"
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/context/content/generic/notification/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockReadRepo struct {
	view  *domain.NotificationView
	views []*domain.NotificationView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id domain.NotificationID) (*domain.NotificationView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrNotificationNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ domain.NotificationFilter) ([]*domain.NotificationView, int64, error) {
	return m.views, m.total, nil
}

type errorReadRepo struct{ err error }

func (m *errorReadRepo) FindByID(_ context.Context, _ domain.NotificationID) (*domain.NotificationView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ domain.NotificationFilter) ([]*domain.NotificationView, int64, error) {
	return nil, 0, m.err
}

var errRepo = errors.New("repo failure")

// --- Tests: GetNotification ---

func TestGetHandler_Handle(t *testing.T) {
	t.Parallel()

	id := domain.NewNotificationID()
	userID := uuid.New()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &domain.NotificationView{
			ID:        id,
			UserID:    userID,
			Title:     "Test Notification",
			Message:   "test message",
			Type:      "INFO",
			CreatedAt: now,
		},
	}

	handler := NewGetHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetQuery{ID: id})
	require.NoError(t, err)
	if result == nil {
		t.Fatal("expected result")
	}
	if result.Title != "Test Notification" {
		t.Errorf("expected title 'Test Notification', got %s", result.Title)
	}
	if result.Type != "INFO" {
		t.Errorf("expected type INFO, got %s", result.Type)
	}
	if result.UserID != userID {
		t.Errorf("expected userID %s, got %s", userID, result.UserID)
	}
}

func TestGetHandler_NotFound(t *testing.T) {
	t.Parallel()

	readRepo := &mockReadRepo{}
	handler := NewGetHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetQuery{ID: domain.NewNotificationID()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepo}
	handler := NewGetHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetQuery{ID: domain.NewNotificationID()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestGetHandler_AllFieldsMapped(t *testing.T) {
	t.Parallel()

	id := domain.NewNotificationID()
	userID := uuid.New()
	now := time.Now()
	readAt := time.Now()

	readRepo := &mockReadRepo{
		view: &domain.NotificationView{
			ID:        id,
			UserID:    userID,
			Title:     "Alert",
			Message:   "Something happened",
			Type:      "WARNING",
			ReadAt:    &readAt,
			CreatedAt: now,
		},
	}

	handler := NewGetHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetQuery{ID: id})
	require.NoError(t, err)
	if result.ReadAt == nil {
		t.Error("readAt not mapped")
	}
	if result.Message != "Something happened" {
		t.Errorf("expected message 'Something happened', got %s", result.Message)
	}
}
