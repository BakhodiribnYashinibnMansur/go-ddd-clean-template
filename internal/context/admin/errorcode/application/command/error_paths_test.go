package command

import (
	"context"
	"errors"
	"testing"

	"gct/internal/context/admin/errorcode/domain"

	"github.com/google/uuid"
)

// --- Error mocks ---

var errRepoSave = errors.New("repo save failed")
var errRepoUpdate = errors.New("repo update failed")
var errRepoDelete = errors.New("repo delete failed")

type errorErrorCodeRepo struct {
	saveErr   error
	updateErr error
	deleteErr error
	findFn    func(ctx context.Context, id uuid.UUID) (*domain.ErrorCode, error)
}

func (m *errorErrorCodeRepo) Save(_ context.Context, _ *domain.ErrorCode) error {
	return m.saveErr
}

func (m *errorErrorCodeRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.ErrorCode, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrErrorCodeNotFound
}

func (m *errorErrorCodeRepo) Update(_ context.Context, _ *domain.ErrorCode) error {
	return m.updateErr
}

func (m *errorErrorCodeRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return m.deleteErr
}

// --- Tests ---

func TestCreateErrorCodeHandler_RepoError(t *testing.T) {
	t.Parallel()

	repo := &errorErrorCodeRepo{saveErr: errRepoSave}
	handler := NewCreateErrorCodeHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), CreateErrorCodeCommand{
		Code: "ERR", Message: "fail", HTTPStatus: 500, Category: "c", Severity: "s",
	})
	if !errors.Is(err, errRepoSave) {
		t.Fatalf("expected errRepoSave, got: %v", err)
	}
}

func TestUpdateErrorCodeHandler_RepoUpdateError(t *testing.T) {
	t.Parallel()

	ec := domain.NewErrorCode("ERR", "msg", 500, "c", "s", false, 0, "")

	repo := &errorErrorCodeRepo{
		findFn:    func(_ context.Context, _ uuid.UUID) (*domain.ErrorCode, error) { return ec, nil },
		updateErr: errRepoUpdate,
	}
	handler := NewUpdateErrorCodeHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), UpdateErrorCodeCommand{
		ID: domain.ErrorCodeID(ec.ID()), Message: "m", HTTPStatus: 500, Category: "c", Severity: "s",
	})
	if !errors.Is(err, errRepoUpdate) {
		t.Fatalf("expected errRepoUpdate, got: %v", err)
	}
}

func TestDeleteErrorCodeHandler_RepoError(t *testing.T) {
	t.Parallel()

	ec := domain.NewErrorCode("ERR_DEL", "msg", 500, "c", "s", false, 0, "")
	repo := &errorErrorCodeRepo{
		findFn:    func(_ context.Context, _ uuid.UUID) (*domain.ErrorCode, error) { return ec, nil },
		deleteErr: errRepoDelete,
	}
	handler := NewDeleteErrorCodeHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), DeleteErrorCodeCommand{ID: domain.ErrorCodeID(ec.ID())})
	if !errors.Is(err, errRepoDelete) {
		t.Fatalf("expected errRepoDelete, got: %v", err)
	}
}

func TestDeleteErrorCodeHandler_FindByIDError(t *testing.T) {
	t.Parallel()

	repo := &errorErrorCodeRepo{} // findFn is nil, returns ErrErrorCodeNotFound
	handler := NewDeleteErrorCodeHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), DeleteErrorCodeCommand{ID: domain.ErrorCodeID(uuid.New())})
	if !errors.Is(err, domain.ErrErrorCodeNotFound) {
		t.Fatalf("expected ErrErrorCodeNotFound, got: %v", err)
	}
}
