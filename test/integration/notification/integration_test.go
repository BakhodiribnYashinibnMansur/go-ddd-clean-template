package notification

import (
	"context"
	"testing"

	"gct/internal/context/content/generic/notification"
	"gct/internal/context/content/generic/notification/application/command"
	"gct/internal/context/content/generic/notification/application/query"
	notifentity "gct/internal/context/content/generic/notification/domain/entity"
	notifrepo "gct/internal/context/content/generic/notification/domain/repository"
	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/logger"
	"gct/test/integration/common/setup"

	"github.com/google/uuid"
)

func newTestBC(t *testing.T) *notification.BoundedContext {
	t.Helper()
	eb := eventbus.NewInMemoryEventBus()
	l := logger.New("error")
	return notification.NewBoundedContext(setup.TestPG.Pool, eb, l)
}

func TestIntegration_CreateAndGetNotification(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	userID := uuid.New()
	err := bc.CreateNotification.Handle(ctx, command.CreateCommand{
		UserID:  userID,
		Title:   "Test Notification",
		Message: "This is a test message",
		Type:    "INFO",
	})
	if err != nil {
		t.Fatalf("CreateNotification: %v", err)
	}

	result, err := bc.ListNotifications.Handle(ctx, query.ListQuery{
		Filter: notifrepo.NotificationFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListNotifications: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 notification, got %d", result.Total)
	}

	n := result.Notifications[0]
	if n.Title != "Test Notification" {
		t.Errorf("expected title Test Notification, got %s", n.Title)
	}
	if n.Message != "This is a test message" {
		t.Errorf("expected message 'This is a test message', got %s", n.Message)
	}

	view, err := bc.GetNotification.Handle(ctx, query.GetQuery{ID: notifentity.NotificationID(n.ID)})
	if err != nil {
		t.Fatalf("GetNotification: %v", err)
	}
	if view.ID != n.ID {
		t.Errorf("ID mismatch: %s vs %s", view.ID, n.ID)
	}
}

func TestIntegration_DeleteNotification(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	userID := uuid.New()
	err := bc.CreateNotification.Handle(ctx, command.CreateCommand{
		UserID:  userID,
		Title:   "To Delete",
		Message: "Will be deleted",
		Type:    "WARNING",
	})
	if err != nil {
		t.Fatalf("CreateNotification: %v", err)
	}

	list, _ := bc.ListNotifications.Handle(ctx, query.ListQuery{
		Filter: notifrepo.NotificationFilter{Limit: 10},
	})
	nID := notifentity.NotificationID(list.Notifications[0].ID)

	err = bc.DeleteNotification.Handle(ctx, command.DeleteCommand{ID: nID})
	if err != nil {
		t.Fatalf("DeleteNotification: %v", err)
	}

	list2, _ := bc.ListNotifications.Handle(ctx, query.ListQuery{
		Filter: notifrepo.NotificationFilter{Limit: 10},
	})
	if list2.Total != 0 {
		t.Errorf("expected 0 notifications after delete, got %d", list2.Total)
	}
}
