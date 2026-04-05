package asynq

import (
	"context"
	"encoding/json"
	"testing"

	"gct/internal/kernel/infrastructure/logger"

	"github.com/hibiken/asynq"
)

func newTestHandlers() *Handlers {
	return NewHandlers(logger.Noop(), nil)
}

func TestHandleImageResize_Success(t *testing.T) {
	h := newTestHandlers()

	payload, err := json.Marshal(ImagePayload{
		SourcePath: "/tmp/source.jpg",
		TargetPath: "/tmp/target.jpg",
		Width:      800,
		Height:     600,
		Quality:    90,
	})
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	task := asynq.NewTask("image:resize", payload)
	if err := h.HandleImageResize(context.Background(), task); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestHandleImageResize_InvalidPayload(t *testing.T) {
	h := newTestHandlers()

	task := asynq.NewTask("image:resize", []byte("invalid json"))
	if err := h.HandleImageResize(context.Background(), task); err == nil {
		t.Fatal("expected error for invalid payload, got nil")
	}
}

func TestHandlePushNotification_Success(t *testing.T) {
	h := newTestHandlers()

	payload, err := json.Marshal(NotificationPayload{
		UserID:  "user-123",
		Title:   "Test",
		Message: "Hello",
		Data:    map[string]string{"key": "value"},
	})
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	task := asynq.NewTask("notification:push", payload)
	if err := h.HandlePushNotification(context.Background(), task); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestHandlePushNotification_InvalidPayload(t *testing.T) {
	h := newTestHandlers()

	task := asynq.NewTask("notification:push", []byte("invalid json"))
	if err := h.HandlePushNotification(context.Background(), task); err == nil {
		t.Fatal("expected error for invalid payload, got nil")
	}
}

func TestHandleAuditLog_InvalidPayload(t *testing.T) {
	h := newTestHandlers()

	task := asynq.NewTask("audit:log", []byte("invalid json"))
	if err := h.HandleAuditLog(context.Background(), task); err == nil {
		t.Fatal("expected error for invalid payload, got nil")
	}
}
