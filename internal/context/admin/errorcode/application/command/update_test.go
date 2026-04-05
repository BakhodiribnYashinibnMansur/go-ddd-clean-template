package command

import (
	"context"
	"testing"

	"gct/internal/context/admin/errorcode/domain"

	"github.com/google/uuid"
)

func TestUpdateErrorCodeHandler_Handle(t *testing.T) {
	ec := domain.NewErrorCode("AUTH_001", "old msg", 401, "auth", "high", false, 0, "old suggestion")

	repo := &mockErrorCodeRepo{
		findFn: func(_ context.Context, id uuid.UUID) (*domain.ErrorCode, error) {
			if id == ec.ID() {
				return ec, nil
			}
			return nil, domain.ErrErrorCodeNotFound
		},
	}
	eb := &mockEventBus{}
	handler := NewUpdateErrorCodeHandler(repo, eb, &mockLogger{})

	cmd := UpdateErrorCodeCommand{
		ID:         ec.ID(),
		Message:    "new msg",
		HTTPStatus: 403,
		Category:   "auth",
		Severity:   "critical",
		Retryable:  true,
		RetryAfter: 30,
		Suggestion: "new suggestion",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
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
	repo := &mockErrorCodeRepo{}
	handler := NewUpdateErrorCodeHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), UpdateErrorCodeCommand{
		ID: uuid.New(), Message: "m", HTTPStatus: 500, Category: "c", Severity: "s",
	})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}
