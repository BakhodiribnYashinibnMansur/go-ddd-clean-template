package domain_test

import (
	"testing"

	"gct/internal/context/content/notification/domain"

	"github.com/google/uuid"
)

func TestNewNotification(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	n, _ := domain.NewNotification(userID, "Welcome", "Hello world", "info")

	if n.UserID() != userID {
		t.Fatalf("expected userID %s, got %s", userID, n.UserID())
	}
	if n.Title() != "Welcome" {
		t.Fatalf("expected title Welcome, got %s", n.Title())
	}
	if n.Message() != "Hello world" {
		t.Fatalf("expected message Hello world, got %s", n.Message())
	}
	if n.Type() != "info" {
		t.Fatalf("expected type info, got %s", n.Type())
	}
	if n.ReadAt() != nil {
		t.Fatal("expected readAt nil")
	}
	if len(n.Events()) != 1 {
		t.Fatalf("expected 1 event, got %d", len(n.Events()))
	}
	if n.Events()[0].EventName() != "notification.sent" {
		t.Fatalf("expected event notification.sent, got %s", n.Events()[0].EventName())
	}
}

func TestNotification_MarkAsRead(t *testing.T) {
	t.Parallel()

	n, _ := domain.NewNotification(uuid.New(), "Test", "msg", "info")

	n.MarkAsRead()
	if n.ReadAt() == nil {
		t.Fatal("expected readAt to be set after MarkAsRead")
	}
}
