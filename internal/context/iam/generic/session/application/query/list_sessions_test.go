package query

import (
	"gct/internal/kernel/infrastructure/logger"
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/context/iam/generic/session/application/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListSessionsHandler_Handle(t *testing.T) {
	t.Parallel()

	now := time.Now()
	readRepo := &mockSessionReadRepository{
		views: []*dto.SessionView{
			{
				ID:           uuid.New(),
				UserID:       uuid.New(),
				DeviceID:     "device-1",
				DeviceName:   "Chrome on Mac",
				DeviceType:   "DESKTOP",
				IPAddress:    "192.168.1.1",
				UserAgent:    "Mozilla/5.0",
				ExpiresAt:    now.Add(7 * 24 * time.Hour),
				LastActivity: now,
				Revoked:      false,
				CreatedAt:    now,
			},
			{
				ID:           uuid.New(),
				UserID:       uuid.New(),
				DeviceID:     "device-2",
				DeviceName:   "Safari on iPhone",
				DeviceType:   "MOBILE",
				IPAddress:    "10.0.0.1",
				UserAgent:    "Safari/16.0",
				ExpiresAt:    now.Add(7 * 24 * time.Hour),
				LastActivity: now,
				Revoked:      false,
				CreatedAt:    now,
			},
		},
		total: 2,
	}

	handler := NewListSessionsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListSessionsQuery{
		Filter: dto.SessionsFilter{Limit: 10, Offset: 0},
	})
	require.NoError(t, err)
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if len(result.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(result.Sessions))
	}
	if result.Sessions[0].DeviceType != "DESKTOP" {
		t.Errorf("expected DESKTOP, got %s", result.Sessions[0].DeviceType)
	}
	if result.Sessions[1].DeviceType != "MOBILE" {
		t.Errorf("expected MOBILE, got %s", result.Sessions[1].DeviceType)
	}
}

func TestListSessionsHandler_Empty(t *testing.T) {
	t.Parallel()

	readRepo := &mockSessionReadRepository{
		views: []*dto.SessionView{},
		total: 0,
	}

	handler := NewListSessionsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListSessionsQuery{
		Filter: dto.SessionsFilter{},
	})
	require.NoError(t, err)
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if len(result.Sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(result.Sessions))
	}
}

func TestListSessionsHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &mockSessionReadRepository{
		err: errors.New("database connection failed"),
	}

	handler := NewListSessionsHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListSessionsQuery{
		Filter: dto.SessionsFilter{},
	})
	if err == nil {
		t.Fatal("expected error when repo fails, got nil")
	}
}
