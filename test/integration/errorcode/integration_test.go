package errorcode

import (
	"context"
	"testing"

	"gct/internal/context/admin/supporting/errorcode"
	"gct/internal/context/admin/supporting/errorcode/application/command"
	"gct/internal/context/admin/supporting/errorcode/application/query"
	errcodeentity "gct/internal/context/admin/supporting/errorcode/domain/entity"
	errcoderepo "gct/internal/context/admin/supporting/errorcode/domain/repository"
	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/logger"
	"gct/test/integration/common/setup"
)

func newTestBC(t *testing.T) *errorcode.BoundedContext {
	t.Helper()
	eb := eventbus.NewInMemoryEventBus()
	l := logger.New("error")
	return errorcode.NewBoundedContext(setup.TestPG.Pool, eb, l)
}

func TestIntegration_CreateAndGetErrorCode(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateErrorCode.Handle(ctx, command.CreateErrorCodeCommand{
		Code:       "AUTH_001",
		Message:    "Unauthorized access",
		HTTPStatus: 401,
		Category:   "AUTH",
		Severity:   "HIGH",
		Retryable:  false,
		RetryAfter: 0,
		Suggestion: "Check your credentials",
	})
	if err != nil {
		t.Fatalf("CreateErrorCode: %v", err)
	}

	result, err := bc.ListErrorCodes.Handle(ctx, query.ListErrorCodesQuery{
		Filter: errcoderepo.ErrorCodeFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListErrorCodes: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 error code, got %d", result.Total)
	}

	a := result.ErrorCodes[0]
	if a.Code != "AUTH_001" {
		t.Errorf("expected code AUTH_001, got %s", a.Code)
	}
	if a.Message != "Unauthorized access" {
		t.Errorf("expected message 'Unauthorized access', got %s", a.Message)
	}

	view, err := bc.GetErrorCode.Handle(ctx, query.GetErrorCodeQuery{ID: errcodeentity.ErrorCodeID(a.ID)})
	if err != nil {
		t.Fatalf("GetErrorCode: %v", err)
	}
	if view.ID != a.ID {
		t.Errorf("ID mismatch: %s vs %s", view.ID, a.ID)
	}
}

func TestIntegration_UpdateErrorCode(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateErrorCode.Handle(ctx, command.CreateErrorCodeCommand{
		Code:       "VAL_001",
		Message:    "Validation failed",
		HTTPStatus: 400,
		Category:   "VALIDATION",
		Severity:   "MEDIUM",
		Retryable:  true,
		RetryAfter: 5,
		Suggestion: "Fix input",
	})
	if err != nil {
		t.Fatalf("CreateErrorCode: %v", err)
	}

	list, _ := bc.ListErrorCodes.Handle(ctx, query.ListErrorCodesQuery{
		Filter: errcoderepo.ErrorCodeFilter{Limit: 10},
	})
	ecID := errcodeentity.ErrorCodeID(list.ErrorCodes[0].ID)

	err = bc.UpdateErrorCode.Handle(ctx, command.UpdateErrorCodeCommand{
		ID:         ecID,
		Message:    "Updated validation message",
		HTTPStatus: 422,
		Category:   "VALIDATION",
		Severity:   "HIGH",
		Retryable:  false,
		RetryAfter: 0,
		Suggestion: "Check input format",
	})
	if err != nil {
		t.Fatalf("UpdateErrorCode: %v", err)
	}

	view, _ := bc.GetErrorCode.Handle(ctx, query.GetErrorCodeQuery{ID: ecID})
	if view.Message != "Updated validation message" {
		t.Errorf("message not updated, got %s", view.Message)
	}
	if view.HTTPStatus != 422 {
		t.Errorf("http_status not updated, got %d", view.HTTPStatus)
	}
}

func TestIntegration_DeleteErrorCode(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateErrorCode.Handle(ctx, command.CreateErrorCodeCommand{
		Code:       "DEL_001",
		Message:    "To be deleted",
		HTTPStatus: 500,
		Category:   "SYSTEM",
		Severity:   "LOW",
		Retryable:  false,
		RetryAfter: 0,
		Suggestion: "Contact support",
	})
	if err != nil {
		t.Fatalf("CreateErrorCode: %v", err)
	}

	list, _ := bc.ListErrorCodes.Handle(ctx, query.ListErrorCodesQuery{
		Filter: errcoderepo.ErrorCodeFilter{Limit: 10},
	})
	ecID := errcodeentity.ErrorCodeID(list.ErrorCodes[0].ID)

	err = bc.DeleteErrorCode.Handle(ctx, command.DeleteErrorCodeCommand{ID: ecID})
	if err != nil {
		t.Fatalf("DeleteErrorCode: %v", err)
	}

	list2, _ := bc.ListErrorCodes.Handle(ctx, query.ListErrorCodesQuery{
		Filter: errcoderepo.ErrorCodeFilter{Limit: 10},
	})
	if list2.Total != 0 {
		t.Errorf("expected 0 error codes after delete, got %d", list2.Total)
	}
}
