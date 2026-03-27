package command_test

import (
	"context"
	"testing"

	"gct/internal/notification/application/command"

	"github.com/google/uuid"
)

func TestDeleteHandler_Handle(t *testing.T) {
	repo := &mockNotificationRepo{}
	handler := command.NewDeleteHandler(repo, &mockEventBus{}, &mockLogger{})

	id := uuid.New()
	cmd := command.DeleteCommand{ID: id}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if repo.deleted != id {
		t.Fatalf("expected deleted ID %s, got %s", id, repo.deleted)
	}
}
