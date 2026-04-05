package command

import (
	"context"
	"errors"
	"testing"

	"gct/internal/context/admin/integration/domain"

	"github.com/google/uuid"
)

// --- Error mocks ---

var errRepoSave = errors.New("repo save failed")
var errRepoUpdate = errors.New("repo update failed")
var errRepoDelete = errors.New("repo delete failed")

type errorIntegrationRepo struct {
	saveErr   error
	updateErr error
	deleteErr error
	findFn    func(ctx context.Context, id uuid.UUID) (*domain.Integration, error)
}

func (m *errorIntegrationRepo) Save(_ context.Context, _ *domain.Integration) error {
	return m.saveErr
}

func (m *errorIntegrationRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Integration, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrIntegrationNotFound
}

func (m *errorIntegrationRepo) Update(_ context.Context, _ *domain.Integration) error {
	return m.updateErr
}

func (m *errorIntegrationRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return m.deleteErr
}

// --- Tests ---

func TestCreateHandler_RepoError(t *testing.T) {
	repo := &errorIntegrationRepo{saveErr: errRepoSave}
	handler := NewCreateHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), CreateCommand{
		Name: "test", Type: "t", APIKey: "k", WebhookURL: "u",
	})
	if !errors.Is(err, errRepoSave) {
		t.Fatalf("expected errRepoSave, got: %v", err)
	}
}

func TestUpdateHandler_RepoUpdateError(t *testing.T) {
	i := domain.NewIntegration("n", "t", "k", "u", true, nil)

	repo := &errorIntegrationRepo{
		findFn:    func(_ context.Context, _ uuid.UUID) (*domain.Integration, error) { return i, nil },
		updateErr: errRepoUpdate,
	}
	handler := NewUpdateHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), UpdateCommand{ID: i.ID()})
	if !errors.Is(err, errRepoUpdate) {
		t.Fatalf("expected errRepoUpdate, got: %v", err)
	}
}

func TestDeleteHandler_RepoError(t *testing.T) {
	repo := &errorIntegrationRepo{deleteErr: errRepoDelete}
	handler := NewDeleteHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), DeleteCommand{ID: uuid.New()})
	if !errors.Is(err, errRepoDelete) {
		t.Fatalf("expected errRepoDelete, got: %v", err)
	}
}
