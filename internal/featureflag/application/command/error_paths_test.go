package command

import (
	"context"
	"errors"
	"testing"

	"gct/internal/featureflag/domain"

	"github.com/google/uuid"
)

// errorRepo is a mock that returns configurable errors for each method.
type errorRepo struct {
	saveErr   error
	updateErr error
	deleteErr error
	findFn    func(ctx context.Context, id uuid.UUID) (*domain.FeatureFlag, error)
}

func (m *errorRepo) Save(_ context.Context, _ *domain.FeatureFlag) error {
	return m.saveErr
}

func (m *errorRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.FeatureFlag, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrFeatureFlagNotFound
}

func (m *errorRepo) Update(_ context.Context, _ *domain.FeatureFlag) error {
	return m.updateErr
}

func (m *errorRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return m.deleteErr
}

// --- Error path tests ---

var errSave = errors.New("save failed")
var errUpdate = errors.New("update failed")
var errDelete = errors.New("delete failed")

func TestCreateHandler_SaveError(t *testing.T) {
	repo := &errorRepo{saveErr: errSave}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateHandler(repo, eb, log)
	err := handler.Handle(context.Background(), CreateCommand{
		Name: "f", Description: "d", Enabled: true, RolloutPercentage: 100,
	})
	if !errors.Is(err, errSave) {
		t.Fatalf("expected errSave, got: %v", err)
	}
}

func TestUpdateHandler_FindError(t *testing.T) {
	repo := &errorRepo{}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateHandler(repo, eb, log)
	newName := "new"
	err := handler.Handle(context.Background(), UpdateCommand{
		ID:   uuid.New(),
		Name: &newName,
	})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestUpdateHandler_UpdateError(t *testing.T) {
	ff := domain.NewFeatureFlag("f", "d", true, 50)

	repo := &errorRepo{
		findFn:    func(_ context.Context, _ uuid.UUID) (*domain.FeatureFlag, error) { return ff, nil },
		updateErr: errUpdate,
	}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateHandler(repo, eb, log)
	newName := "updated"
	err := handler.Handle(context.Background(), UpdateCommand{
		ID:   ff.ID(),
		Name: &newName,
	})
	if !errors.Is(err, errUpdate) {
		t.Fatalf("expected errUpdate, got: %v", err)
	}
}

func TestDeleteHandler_DeleteError(t *testing.T) {
	repo := &errorRepo{deleteErr: errDelete}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewDeleteHandler(repo, eb, log)
	err := handler.Handle(context.Background(), DeleteCommand{ID: uuid.New()})
	if !errors.Is(err, errDelete) {
		t.Fatalf("expected errDelete, got: %v", err)
	}
}
