package command

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestDeleteHandler_Handle(t *testing.T) {
	repo := &mockFeatureFlagRepo{}
	eb := &mockEventBus{}
	handler := NewDeleteHandler(repo, eb, &mockLogger{})

	id := uuid.New()
	err := handler.Handle(context.Background(), DeleteCommand{ID: id})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

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
	repoErr := errors.New("delete failed")
	repo := &mockFeatureFlagRepo{
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return repoErr
		},
	}
	handler := NewDeleteHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), DeleteCommand{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got: %v", err)
	}
}
