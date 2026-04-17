package command

import (
	"context"
	"errors"
	"testing"

	ffentity "gct/internal/context/admin/generic/featureflag/domain/entity"
	shareddomain "gct/internal/kernel/domain"

	"gct/internal/kernel/outbox"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDeleteHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockFeatureFlagRepo{}
	eb := &mockEventBus{}
	handler := NewDeleteHandler(repo, outbox.NewEventCommitter(nil, nil, eb, &mockLogger{}), &mockLogger{})

	id := ffentity.NewFeatureFlagID()
	err := handler.Handle(context.Background(), DeleteCommand{ID: ffentity.FeatureFlagID(id)})
	require.NoError(t, err)

	if repo.deleted != id {
		t.Errorf("expected deleted ID %s, got %s", id, repo.deleted)
	}
	if len(eb.published) == 0 {
		t.Error("expected FlagDeleted event to be published")
	}
	if eb.published[0].EventName() != "featureflag.deleted" {
		t.Errorf("expected event name featureflag.deleted, got %s", eb.published[0].EventName())
	}
}

func TestDeleteHandler_Handle_RepoError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("delete failed")
	repo := &mockFeatureFlagRepo{
		deleteFn: func(_ context.Context, _ shareddomain.Querier, _ ffentity.FeatureFlagID) error {
			return repoErr
		},
	}
	handler := NewDeleteHandler(repo, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})

	err := handler.Handle(context.Background(), DeleteCommand{ID: ffentity.FeatureFlagID(uuid.New())})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got: %v", err)
	}
}
