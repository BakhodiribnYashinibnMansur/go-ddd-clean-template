package command

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestDeleteRoleHandler_Success(t *testing.T) {
	repo := &mockRoleRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewDeleteRoleHandler(repo, eventBus, log)

	roleID := uuid.New()
	cmd := DeleteRoleCommand{ID: roleID}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(eventBus.publishedEvents) == 0 {
		t.Fatal("expected at least one event to be published")
	}

	if eventBus.publishedEvents[0].EventName() != "authz.role_deleted" {
		t.Errorf("expected event authz.role_deleted, got %s", eventBus.publishedEvents[0].EventName())
	}

	if eventBus.publishedEvents[0].AggregateID() != roleID {
		t.Errorf("expected aggregate ID %s, got %s", roleID, eventBus.publishedEvents[0].AggregateID())
	}
}
