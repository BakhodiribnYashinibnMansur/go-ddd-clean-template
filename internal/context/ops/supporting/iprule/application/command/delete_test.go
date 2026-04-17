package command

import (
	"context"
	"testing"

	ipruleentity "gct/internal/context/ops/supporting/iprule/domain/entity"
	"gct/internal/kernel/outbox"
	"github.com/stretchr/testify/require"
)

func TestDeleteIPRuleHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockIPRuleRepo{}
	handler := NewDeleteIPRuleHandler(repo, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})

	id := ipruleentity.NewIPRuleID()
	err := handler.Handle(context.Background(), DeleteIPRuleCommand{ID: id})
	require.NoError(t, err)
	if repo.deleted != id {
		t.Errorf("expected deleted ID %s, got %s", id, repo.deleted)
	}
}
