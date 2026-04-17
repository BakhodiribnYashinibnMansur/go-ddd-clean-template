package command

import (
	"context"
	"testing"

	announceentity "gct/internal/context/content/supporting/announcement/domain/entity"

	"gct/internal/kernel/outbox"
	"github.com/stretchr/testify/require"
)

func TestDeleteAnnouncementHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockAnnouncementRepo{}
	handler := NewDeleteAnnouncementHandler(repo, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})

	id := announceentity.NewAnnouncementID()
	err := handler.Handle(context.Background(), DeleteAnnouncementCommand{ID: id})
	require.NoError(t, err)
	if repo.deleted != id {
		t.Errorf("expected deleted ID %s, got %s", id, repo.deleted)
	}
}
