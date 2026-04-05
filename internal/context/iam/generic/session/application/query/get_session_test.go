package query

import (
	"gct/internal/kernel/infrastructure/logger"
	"context"
	"errors"
	"testing"
	"time"

	appdto "gct/internal/context/iam/generic/session/application"
	sessiondomain "gct/internal/context/iam/generic/session/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mock Read Repository ---

type mockSessionReadRepository struct {
	view  *appdto.SessionView
	views []*appdto.SessionView
	total int64
	err   error
}

func (m *mockSessionReadRepository) FindByID(_ context.Context, id uuid.UUID) (*appdto.SessionView, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, errors.New("session not found")
}

func (m *mockSessionReadRepository) List(_ context.Context, _ appdto.SessionsFilter) ([]*appdto.SessionView, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	return m.views, m.total, nil
}

// --- Mock Logger ---

type mockLogger struct{}

func (ml *mockLogger) Debug(_ ...any)                                       {}
func (ml *mockLogger) Debugf(_ string, _ ...any)                            {}
func (ml *mockLogger) Debugw(_ string, _ ...any)                            {}
func (ml *mockLogger) Info(_ ...any)                                        {}
func (ml *mockLogger) Infof(_ string, _ ...any)                             {}
func (ml *mockLogger) Infow(_ string, _ ...any)                             {}
func (ml *mockLogger) Warn(_ ...any)                                        {}
func (ml *mockLogger) Warnf(_ string, _ ...any)                             {}
func (ml *mockLogger) Warnw(_ string, _ ...any)                             {}
func (ml *mockLogger) Error(_ ...any)                                       {}
func (ml *mockLogger) Errorf(_ string, _ ...any)                            {}
func (ml *mockLogger) Errorw(_ string, _ ...any)                            {}
func (ml *mockLogger) Fatal(_ ...any)                                       {}
func (ml *mockLogger) Fatalf(_ string, _ ...any)                            {}
func (ml *mockLogger) Fatalw(_ string, _ ...any)                            {}
func (ml *mockLogger) Debugc(_ context.Context, _ string, _ ...any)         {}
func (ml *mockLogger) Infoc(_ context.Context, _ string, _ ...any)          {}
func (ml *mockLogger) Warnc(_ context.Context, _ string, _ ...any)          {}
func (ml *mockLogger) Errorc(_ context.Context, _ string, _ ...any)         {}
func (ml *mockLogger) Fatalc(_ context.Context, _ string, _ ...any)         {}

// --- Tests ---

func TestGetSessionHandler_Handle(t *testing.T) {
	t.Parallel()

	sessionID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	readRepo := &mockSessionReadRepository{
		view: &appdto.SessionView{
			ID:           sessionID,
			UserID:       userID,
			DeviceID:     "device-123",
			DeviceName:   "Chrome on Mac",
			DeviceType:   "DESKTOP",
			IPAddress:    "192.168.1.1",
			UserAgent:    "Mozilla/5.0",
			ExpiresAt:    now.Add(7 * 24 * time.Hour),
			LastActivity: now,
			Revoked:      false,
			CreatedAt:    now,
		},
	}

	handler := NewGetSessionHandler(readRepo, logger.Noop())

	q := GetSessionQuery{ID: sessiondomain.SessionID(sessionID)}
	result, err := handler.Handle(context.Background(), q)
	require.NoError(t, err)

	if result == nil {
		t.Fatal("expected session view, got nil")
	}

	if result.ID != sessionID {
		t.Errorf("expected ID %s, got %s", sessionID, result.ID)
	}

	if result.UserID != userID {
		t.Errorf("expected UserID %s, got %s", userID, result.UserID)
	}

	if result.DeviceType != "DESKTOP" {
		t.Errorf("expected DeviceType DESKTOP, got %s", result.DeviceType)
	}

	if result.IPAddress != "192.168.1.1" {
		t.Errorf("expected IPAddress 192.168.1.1, got %s", result.IPAddress)
	}

	if result.Revoked {
		t.Error("expected session to not be revoked")
	}
}

func TestGetSessionHandler_NotFound(t *testing.T) {
	t.Parallel()

	readRepo := &mockSessionReadRepository{}

	handler := NewGetSessionHandler(readRepo, logger.Noop())

	q := GetSessionQuery{ID: sessiondomain.NewSessionID()}
	_, err := handler.Handle(context.Background(), q)
	if err == nil {
		t.Fatal("expected error for non-existent session, got nil")
	}
}

func TestGetSessionHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &mockSessionReadRepository{
		err: errors.New("database connection failed"),
	}

	handler := NewGetSessionHandler(readRepo, logger.Noop())

	q := GetSessionQuery{ID: sessiondomain.NewSessionID()}
	_, err := handler.Handle(context.Background(), q)
	if err == nil {
		t.Fatal("expected error when repo fails, got nil")
	}
}
