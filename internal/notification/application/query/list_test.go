package query

import (
	"context"
	"testing"
	"time"

	"gct/internal/notification/domain"

	"github.com/google/uuid"
)

func TestListHandler_Handle(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*domain.NotificationView{
			{ID: uuid.New(), UserID: uuid.New(), Title: "N1", Type: "INFO", CreatedAt: time.Now()},
			{ID: uuid.New(), UserID: uuid.New(), Title: "N2", Type: "WARNING", CreatedAt: time.Now()},
		},
		total: 2,
	}

	handler := NewListHandler(readRepo)
	result, err := handler.Handle(context.Background(), ListQuery{
		Filter: domain.NotificationFilter{Limit: 10, Offset: 0},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if len(result.Notifications) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(result.Notifications))
	}
	if result.Notifications[0].Title != "N1" {
		t.Errorf("expected N1, got %s", result.Notifications[0].Title)
	}
}

func TestListHandler_Empty(t *testing.T) {
	readRepo := &mockReadRepo{views: []*domain.NotificationView{}, total: 0}

	handler := NewListHandler(readRepo)
	result, err := handler.Handle(context.Background(), ListQuery{
		Filter: domain.NotificationFilter{},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if len(result.Notifications) != 0 {
		t.Errorf("expected 0 notifications, got %d", len(result.Notifications))
	}
}

func TestListHandler_RepoError(t *testing.T) {
	readRepo := &errorReadRepo{err: errRepo}
	handler := NewListHandler(readRepo)
	_, err := handler.Handle(context.Background(), ListQuery{Filter: domain.NotificationFilter{}})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
