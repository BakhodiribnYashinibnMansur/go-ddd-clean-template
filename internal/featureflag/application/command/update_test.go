package command

import (
	"context"
	"errors"
	"testing"

	"gct/internal/featureflag/domain"

	"github.com/google/uuid"
)

func TestUpdateHandler_Handle(t *testing.T) {
	ff := domain.NewFeatureFlag("old_flag", "old desc", false, 10)

	repo := &mockRepo{
		findFn: func(_ context.Context, id uuid.UUID) (*domain.FeatureFlag, error) {
			if id == ff.ID() {
				return ff, nil
			}
			return nil, domain.ErrFeatureFlagNotFound
		},
	}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateHandler(repo, eb, log)

	newName := "new_flag"
	enabled := true
	rollout := 75
	cmd := UpdateCommand{
		ID:                ff.ID(),
		Name:              &newName,
		Enabled:           &enabled,
		RolloutPercentage: &rollout,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.updated == nil {
		t.Fatal("expected feature flag to be updated")
	}
	if repo.updated.Name() != "new_flag" {
		t.Errorf("expected name new_flag, got %s", repo.updated.Name())
	}
	if repo.updated.Enabled() != true {
		t.Errorf("expected enabled true, got %v", repo.updated.Enabled())
	}
	if repo.updated.RolloutPercentage() != 75 {
		t.Errorf("expected rollout 75, got %d", repo.updated.RolloutPercentage())
	}
	// Unchanged fields should be preserved
	if repo.updated.Description() != "old desc" {
		t.Errorf("expected description old desc (unchanged), got %s", repo.updated.Description())
	}
}

func TestUpdateHandler_NotFound(t *testing.T) {
	repo := &mockRepo{}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateHandler(repo, eb, log)

	newName := "name"
	err := handler.Handle(context.Background(), UpdateCommand{
		ID:   uuid.New(),
		Name: &newName,
	})
	if err == nil {
		t.Fatal("expected error for non-existent feature flag")
	}
}

func TestUpdateHandler_RepoUpdateError(t *testing.T) {
	ff := domain.NewFeatureFlag("f", "d", true, 100)
	repoErr := errors.New("repo update failed")

	errR := &errorRepo{
		findFn:    func(_ context.Context, _ uuid.UUID) (*domain.FeatureFlag, error) { return ff, nil },
		updateErr: repoErr,
	}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateHandler(errR, eb, log)

	newName := "new"
	err := handler.Handle(context.Background(), UpdateCommand{
		ID:   ff.ID(),
		Name: &newName,
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo update error, got: %v", err)
	}
}
