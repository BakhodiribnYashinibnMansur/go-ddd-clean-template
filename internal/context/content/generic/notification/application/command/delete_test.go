package command_test

import (
	"context"
	"testing"

	"gct/internal/context/content/generic/notification/application/command"
	"gct/internal/context/content/generic/notification/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDeleteHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockNotificationRepo{}
	handler := command.NewDeleteHandler(repo, &mockEventBus{}, &mockLogger{})

	id := uuid.New()
	cmd := command.DeleteCommand{ID: domain.NotificationID(id)}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.deleted != id {
		t.Fatalf("expected deleted ID %s, got %s", id, repo.deleted)
	}
}
