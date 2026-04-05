package command

import (
	"context"
	"errors"
	"testing"

	"gct/internal/context/content/translation/domain"

	"github.com/google/uuid"
)

// errorRepo is a mock that returns configurable errors for each method.
type errorRepo struct {
	saveErr   error
	updateErr error
	deleteErr error
	findFn    func(ctx context.Context, id uuid.UUID) (*domain.Translation, error)
}

func (m *errorRepo) Save(_ context.Context, _ *domain.Translation) error {
	return m.saveErr
}

func (m *errorRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Translation, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrTranslationNotFound
}

func (m *errorRepo) Update(_ context.Context, _ *domain.Translation) error {
	return m.updateErr
}

func (m *errorRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return m.deleteErr
}

func (m *errorRepo) List(_ context.Context, _ domain.TranslationFilter) ([]*domain.Translation, int64, error) {
	return nil, 0, nil
}

// --- Error path tests ---

var errSave = errors.New("save failed")
var errUpdate = errors.New("update failed")
var errDelete = errors.New("delete failed")

func TestCreateTranslationHandler_SaveError(t *testing.T) {
	repo := &errorRepo{saveErr: errSave}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateTranslationHandler(repo, eb, log)
	err := handler.Handle(context.Background(), CreateTranslationCommand{
		Key: "k", Language: "en", Value: "v", Group: "g",
	})
	if !errors.Is(err, errSave) {
		t.Fatalf("expected errSave, got: %v", err)
	}
}

func TestUpdateTranslationHandler_FindError(t *testing.T) {
	repo := &errorRepo{}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateTranslationHandler(repo, eb, log)
	newVal := "new"
	err := handler.Handle(context.Background(), UpdateTranslationCommand{
		ID:    uuid.New(),
		Value: &newVal,
	})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestUpdateTranslationHandler_UpdateError(t *testing.T) {
	tr := domain.NewTranslation("k", "en", "v", "g")

	repo := &errorRepo{
		findFn:    func(_ context.Context, _ uuid.UUID) (*domain.Translation, error) { return tr, nil },
		updateErr: errUpdate,
	}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateTranslationHandler(repo, eb, log)
	newVal := "updated"
	err := handler.Handle(context.Background(), UpdateTranslationCommand{
		ID:    tr.ID(),
		Value: &newVal,
	})
	if !errors.Is(err, errUpdate) {
		t.Fatalf("expected errUpdate, got: %v", err)
	}
}

func TestDeleteTranslationHandler_DeleteError(t *testing.T) {
	repo := &errorRepo{deleteErr: errDelete}
	log := &mockLogger{}

	handler := NewDeleteTranslationHandler(repo, log)
	err := handler.Handle(context.Background(), DeleteTranslationCommand{ID: uuid.New()})
	if !errors.Is(err, errDelete) {
		t.Fatalf("expected errDelete, got: %v", err)
	}
}
