package command

import (
	"context"
	"testing"

	"gct/internal/context/content/announcement/domain"
	"github.com/stretchr/testify/require"
)

func TestDeleteAnnouncementHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockAnnouncementRepo{}
	handler := NewDeleteAnnouncementHandler(repo, &mockLogger{})

	id := domain.NewAnnouncementID()
	err := handler.Handle(context.Background(), DeleteAnnouncementCommand{ID: id})
	require.NoError(t, err)
	if repo.deleted != id.UUID() {
		t.Errorf("expected deleted ID %s, got %s", id, repo.deleted)
	}
}
