package command

import (
	"context"
	"testing"

	"gct/internal/context/ops/iprule/domain"
	"github.com/stretchr/testify/require"
)

func TestDeleteIPRuleHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockIPRuleRepo{}
	handler := NewDeleteIPRuleHandler(repo, &mockLogger{})

	id := domain.NewIPRuleID()
	err := handler.Handle(context.Background(), DeleteIPRuleCommand{ID: id})
	require.NoError(t, err)
	if repo.deleted != id.UUID() {
		t.Errorf("expected deleted ID %s, got %s", id, repo.deleted)
	}
}
