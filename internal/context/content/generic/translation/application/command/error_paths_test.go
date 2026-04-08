package command

import (
	"context"
	"errors"
	"testing"

	translationentity "gct/internal/context/content/generic/translation/domain/entity"
	translationrepo "gct/internal/context/content/generic/translation/domain/repository"
	"gct/internal/kernel/outbox"
)

// errorRepo is a mock that returns configurable errors for each method.
type errorRepo struct {
	saveErr   error
	updateErr error
	deleteErr error
	findFn    func(ctx context.Context, id translationentity.TranslationID) (*translationentity.Translation, error)
}

func (m *errorRepo) Save(_ context.Context, _ *translationentity.Translation) error {
	return m.saveErr
}

func (m *errorRepo) FindByID(ctx context.Context, id translationentity.TranslationID) (*translationentity.Translation, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, translationentity.ErrTranslationNotFound
}

func (m *errorRepo) Update(_ context.Context, _ *translationentity.Translation) error {
	return m.updateErr
}

func (m *errorRepo) Delete(_ context.Context, _ translationentity.TranslationID) error {
	return m.deleteErr
}

func (m *errorRepo) List(_ context.Context, _ translationrepo.TranslationFilter) ([]*translationentity.Translation, int64, error) {
	return nil, 0, nil
}

// --- Error path tests ---

var errSave = errors.New("save failed")
var errUpdate = errors.New("update failed")
var errDelete = errors.New("delete failed")

func TestCreateTranslationHandler_SaveError(t *testing.T) {
	t.Parallel()

	repo := &errorRepo{saveErr: errSave}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateTranslationHandler(repo, outbox.NewEventCommitter(nil, nil, eb, log), log)
	err := handler.Handle(context.Background(), CreateTranslationCommand{
		Key: "k", Language: "en", Value: "v", Group: "g",
	})
	if !errors.Is(err, errSave) {
		t.Fatalf("expected errSave, got: %v", err)
	}
}

func TestUpdateTranslationHandler_FindError(t *testing.T) {
	t.Parallel()

	repo := &errorRepo{}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateTranslationHandler(repo, outbox.NewEventCommitter(nil, nil, eb, log), log)
	newVal := "new"
	err := handler.Handle(context.Background(), UpdateTranslationCommand{
		ID:    translationentity.NewTranslationID(),
		Value: &newVal,
	})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestUpdateTranslationHandler_UpdateError(t *testing.T) {
	t.Parallel()

	tr := translationentity.NewTranslation("k", "en", "v", "g")

	repo := &errorRepo{
		findFn:    func(_ context.Context, _ translationentity.TranslationID) (*translationentity.Translation, error) { return tr, nil },
		updateErr: errUpdate,
	}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateTranslationHandler(repo, outbox.NewEventCommitter(nil, nil, eb, log), log)
	newVal := "updated"
	err := handler.Handle(context.Background(), UpdateTranslationCommand{
		ID:    translationentity.TranslationID(tr.ID()),
		Value: &newVal,
	})
	if !errors.Is(err, errUpdate) {
		t.Fatalf("expected errUpdate, got: %v", err)
	}
}

func TestDeleteTranslationHandler_DeleteError(t *testing.T) {
	t.Parallel()

	repo := &errorRepo{deleteErr: errDelete}
	log := &mockLogger{}

	handler := NewDeleteTranslationHandler(repo, log)
	err := handler.Handle(context.Background(), DeleteTranslationCommand{ID: translationentity.NewTranslationID()})
	if !errors.Is(err, errDelete) {
		t.Fatalf("expected errDelete, got: %v", err)
	}
}
