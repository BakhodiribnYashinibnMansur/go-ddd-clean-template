package command

import (
	"context"
	"errors"
	"testing"

	translationentity "gct/internal/context/content/generic/translation/domain/entity"

	"github.com/stretchr/testify/require"
)

func TestUpdateTranslationHandler_Handle(t *testing.T) {
	t.Parallel()

	tr := translationentity.NewTranslation("old_key", "en", "Old Value", "general")

	repo := &mockRepo{
		findFn: func(_ context.Context, id translationentity.TranslationID) (*translationentity.Translation, error) {
			if id == tr.TypedID() {
				return tr, nil
			}
			return nil, translationentity.ErrTranslationNotFound
		},
	}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateTranslationHandler(repo, eb, log)

	newKey := "new_key"
	newValue := "New Value"
	cmd := UpdateTranslationCommand{
		ID:    translationentity.TranslationID(tr.ID()),
		Key:   &newKey,
		Value: &newValue,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

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
	t.Parallel()

	repo := &mockRepo{}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateTranslationHandler(repo, eb, log)

	newKey := "k"
	err := handler.Handle(context.Background(), UpdateTranslationCommand{
		ID:  translationentity.NewTranslationID(),
		Key: &newKey,
	})
	if err == nil {
		t.Fatal("expected error for non-existent translation")
	}
}

func TestUpdateTranslationHandler_RepoUpdateError(t *testing.T) {
	t.Parallel()

	tr := translationentity.NewTranslation("k", "en", "v", "g")
	repoErr := errors.New("repo update failed")

	errR := &errorRepo{
		findFn:    func(_ context.Context, _ translationentity.TranslationID) (*translationentity.Translation, error) { return tr, nil },
		updateErr: repoErr,
	}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateTranslationHandler(errR, eb, log)

	newVal := "new"
	err := handler.Handle(context.Background(), UpdateTranslationCommand{
		ID:    translationentity.TranslationID(tr.ID()),
		Value: &newVal,
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo update error, got: %v", err)
	}
}
