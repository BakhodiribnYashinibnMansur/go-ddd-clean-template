package command

import (
	"context"
	"testing"

	integentity "gct/internal/context/admin/supporting/integration/domain/entity"

	"github.com/stretchr/testify/require"
)

func TestDeleteHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockIntegrationRepo{}
	handler := NewDeleteHandler(repo, &mockEventBus{}, &mockLogger{})

	id := integentity.NewIntegrationID()
	err := handler.Handle(context.Background(), DeleteCommand{ID: id})
	require.NoError(t, err)
	if repo.deleted != id {
		t.Errorf("expected deleted ID %s, got %s", id, repo.deleted)
	}
}
