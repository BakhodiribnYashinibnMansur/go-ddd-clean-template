package command

import (
	"context"
	"testing"

	"gct/internal/context/admin/integration/domain"
	"github.com/stretchr/testify/require"
)

func TestDeleteHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockIntegrationRepo{}
	handler := NewDeleteHandler(repo, &mockEventBus{}, &mockLogger{})

	id := domain.NewIntegrationID()
	err := handler.Handle(context.Background(), DeleteCommand{ID: id})
	require.NoError(t, err)
	if repo.deleted != id.UUID() {
		t.Errorf("expected deleted ID %s, got %s", id, repo.deleted)
	}
}
