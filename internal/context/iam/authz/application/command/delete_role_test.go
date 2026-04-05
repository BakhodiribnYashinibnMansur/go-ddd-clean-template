package command

import (
	"context"
	"testing"

	"gct/internal/context/iam/authz/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDeleteRoleHandler_Success(t *testing.T) {
	t.Parallel()

	repo := &mockRoleRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewDeleteRoleHandler(repo, eventBus, log)

	roleID := uuid.New()
	cmd := DeleteRoleCommand{ID: domain.RoleID(roleID)}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

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
