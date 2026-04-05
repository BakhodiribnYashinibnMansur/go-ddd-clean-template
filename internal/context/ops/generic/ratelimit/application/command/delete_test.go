package command

import (
	"context"
	"testing"

	"gct/internal/context/ops/generic/ratelimit/domain"

	"github.com/stretchr/testify/require"
)

func TestDeleteRateLimitHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockRateLimitRepo{}
	handler := NewDeleteRateLimitHandler(repo, &mockLogger{})

	id := domain.NewRateLimitID()
	err := handler.Handle(context.Background(), DeleteRateLimitCommand{ID: id})
	require.NoError(t, err)
	if repo.deleted != id {
		t.Errorf("expected deleted ID %s, got %s", id, repo.deleted)
	}
}
