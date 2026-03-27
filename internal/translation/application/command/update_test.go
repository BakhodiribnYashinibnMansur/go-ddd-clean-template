package command

import (
	"context"
	"errors"
	"testing"

	"gct/internal/translation/domain"

	"github.com/google/uuid"
)

func TestUpdateTranslationHandler_Handle(t *testing.T) {
	tr := domain.NewTranslation("old_key", "en", "Old Value", "general")

	repo := &mockRepo{
		findFn: func(_ context.Context, id uuid.UUID) (*domain.Translation, error) {
			if id == tr.ID() {
				return tr, nil
			}
			return nil, domain.ErrTranslationNotFound
		},
	}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateTranslationHandler(repo, eb, log)

	newKey := "new_key"
	newValue := "New Value"
	cmd := UpdateTranslationCommand{
		ID:    tr.ID(),
		Key:   &newKey,
		Value: &newValue,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.updated == nil {
		t.Fatal("expected translation to be updated")
	}
	if repo.updated.Key() != "new_key" {
		t.Errorf("expected key new_key, got %s", repo.updated.Key())
	}
	if repo.updated.Value() != "New Value" {
		t.Errorf("expected value New Value, got %s", repo.updated.Value())
	}
	// Unchanged fields should be preserved
	if repo.updated.Language() != "en" {
		t.Errorf("expected language en (unchanged), got %s", repo.updated.Language())
	}
	if repo.updated.Group() != "general" {
		t.Errorf("expected group general (unchanged), got %s", repo.updated.Group())
	}

	if len(eb.published) == 0 {
		t.Fatal("expected events to be published")
	}
	if eb.published[0].EventName() != "translation.updated" {
		t.Errorf("expected translation.updated event, got %s", eb.published[0].EventName())
	}
}

func TestUpdateTranslationHandler_NotFound(t *testing.T) {
	repo := &mockRepo{}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateTranslationHandler(repo, eb, log)

	newKey := "k"
	err := handler.Handle(context.Background(), UpdateTranslationCommand{
		ID:  uuid.New(),
		Key: &newKey,
	})
	if err == nil {
		t.Fatal("expected error for non-existent translation")
	}
}

func TestUpdateTranslationHandler_RepoUpdateError(t *testing.T) {
	tr := domain.NewTranslation("k", "en", "v", "g")
	repoErr := errors.New("repo update failed")

	errR := &errorRepo{
		findFn:    func(_ context.Context, _ uuid.UUID) (*domain.Translation, error) { return tr, nil },
		updateErr: repoErr,
	}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateTranslationHandler(errR, eb, log)

	newVal := "new"
	err := handler.Handle(context.Background(), UpdateTranslationCommand{
		ID:    tr.ID(),
		Value: &newVal,
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo update error, got: %v", err)
	}
}
