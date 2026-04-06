package command

import (
	"context"
	"errors"
	"testing"

	errcodeentity "gct/internal/context/admin/supporting/errorcode/domain/entity"

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
	findFn    func(ctx context.Context, id errcodeentity.ErrorCodeID) (*errcodeentity.ErrorCode, error)
}

func (m *errorErrorCodeRepo) Save(_ context.Context, _ *errcodeentity.ErrorCode) error {
	return m.saveErr
}

func (m *errorErrorCodeRepo) FindByID(ctx context.Context, id errcodeentity.ErrorCodeID) (*errcodeentity.ErrorCode, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, errcodeentity.ErrErrorCodeNotFound
}

func (m *errorErrorCodeRepo) Update(_ context.Context, _ *errcodeentity.ErrorCode) error {
	return m.updateErr
}

func (m *errorErrorCodeRepo) Delete(_ context.Context, _ errcodeentity.ErrorCodeID) error {
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

	ec := errcodeentity.NewErrorCode("ERR", "msg", 500, "c", "s", false, 0, "")

	repo := &errorErrorCodeRepo{
		findFn:    func(_ context.Context, _ errcodeentity.ErrorCodeID) (*errcodeentity.ErrorCode, error) { return ec, nil },
		updateErr: errRepoUpdate,
	}
	handler := NewUpdateErrorCodeHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), UpdateErrorCodeCommand{
		ID: errcodeentity.ErrorCodeID(ec.ID()), Message: "m", HTTPStatus: 500, Category: "c", Severity: "s",
	})
	if !errors.Is(err, errRepoUpdate) {
		t.Fatalf("expected errRepoUpdate, got: %v", err)
	}
}

func TestDeleteErrorCodeHandler_RepoError(t *testing.T) {
	t.Parallel()

	ec := errcodeentity.NewErrorCode("ERR_DEL", "msg", 500, "c", "s", false, 0, "")
	repo := &errorErrorCodeRepo{
		findFn:    func(_ context.Context, _ errcodeentity.ErrorCodeID) (*errcodeentity.ErrorCode, error) { return ec, nil },
		deleteErr: errRepoDelete,
	}
	handler := NewDeleteErrorCodeHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), DeleteErrorCodeCommand{ID: errcodeentity.ErrorCodeID(ec.ID())})
	if !errors.Is(err, errRepoDelete) {
		t.Fatalf("expected errRepoDelete, got: %v", err)
	}
}

func TestDeleteErrorCodeHandler_FindByIDError(t *testing.T) {
	t.Parallel()

	repo := &errorErrorCodeRepo{} // findFn is nil, returns ErrErrorCodeNotFound
	handler := NewDeleteErrorCodeHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), DeleteErrorCodeCommand{ID: errcodeentity.ErrorCodeID(uuid.New())})
	if !errors.Is(err, errcodeentity.ErrErrorCodeNotFound) {
		t.Fatalf("expected ErrErrorCodeNotFound, got: %v", err)
	}
}
