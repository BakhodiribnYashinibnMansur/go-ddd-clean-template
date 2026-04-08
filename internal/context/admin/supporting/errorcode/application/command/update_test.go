package command

import (
	"context"
	"testing"

	errcodeentity "gct/internal/context/admin/supporting/errorcode/domain/entity"

	"gct/internal/kernel/outbox"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUpdateErrorCodeHandler_Handle(t *testing.T) {
	t.Parallel()

	ec := errcodeentity.NewErrorCode("AUTH_001", "old msg", 401, "auth", "high", false, 0, "old suggestion")

	repo := &mockErrorCodeRepo{
		findFn: func(_ context.Context, id errcodeentity.ErrorCodeID) (*errcodeentity.ErrorCode, error) {
			if id == ec.TypedID() {
				return ec, nil
			}
			return nil, errcodeentity.ErrErrorCodeNotFound
		},
	}
	eb := &mockEventBus{}
	handler := NewUpdateErrorCodeHandler(repo, outbox.NewEventCommitter(nil, nil, eb, &mockLogger{}), &mockLogger{})

	cmd := UpdateErrorCodeCommand{
		ID:         errcodeentity.ErrorCodeID(ec.ID()),
		Message:    "new msg",
		HTTPStatus: 403,
		Category:   "auth",
		Severity:   "critical",
		Retryable:  true,
		RetryAfter: 30,
		Suggestion: "new suggestion",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)
	if repo.updated == nil {
		t.Fatal("expected error code to be updated")
	}
	if repo.updated.Message() != "new msg" {
		t.Errorf("expected message 'new msg', got %s", repo.updated.Message())
	}
	if repo.updated.HTTPStatus() != 403 {
		t.Errorf("expected httpStatus 403, got %d", repo.updated.HTTPStatus())
	}
	if repo.updated.Severity() != "critical" {
		t.Errorf("expected severity critical, got %s", repo.updated.Severity())
	}
	if !repo.updated.Retryable() {
		t.Error("expected retryable true")
	}
	if repo.updated.RetryAfter() != 30 {
		t.Errorf("expected retryAfter 30, got %d", repo.updated.RetryAfter())
	}

	// Code should remain immutable
	if repo.updated.Code() != "AUTH_001" {
		t.Errorf("expected code AUTH_001 (immutable), got %s", repo.updated.Code())
	}

	if len(eb.published) == 0 {
		t.Fatal("expected events to be published")
	}
}

func TestUpdateErrorCodeHandler_NotFound(t *testing.T) {
	t.Parallel()

	repo := &mockErrorCodeRepo{}
	handler := NewUpdateErrorCodeHandler(repo, outbox.NewEventCommitter(nil, nil, &mockEventBus{}, &mockLogger{}), &mockLogger{})

	err := handler.Handle(context.Background(), UpdateErrorCodeCommand{
		ID: errcodeentity.ErrorCodeID(uuid.New()), Message: "m", HTTPStatus: 500, Category: "c", Severity: "s",
	})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}
