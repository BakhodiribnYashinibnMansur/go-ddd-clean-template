package systemerror

import (
	"context"
	"testing"

	"gct/internal/platform/infrastructure/eventbus"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/context/ops/systemerror"
	"gct/internal/context/ops/systemerror/application/command"
	"gct/internal/context/ops/systemerror/application/query"
	"gct/internal/context/ops/systemerror/domain"
	"gct/test/integration/common/setup"

	"github.com/google/uuid"
)

func newTestBC(t *testing.T) *systemerror.BoundedContext {
	t.Helper()
	eb := eventbus.NewInMemoryEventBus()
	l := logger.New("error")
	return systemerror.NewBoundedContext(setup.TestPG.Pool, eb, l)
}

func TestIntegration_CreateAndGetSystemError(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	stackTrace := "goroutine 1 [running]:\nmain.main()"
	err := bc.CreateSystemError.Handle(ctx, command.CreateSystemErrorCommand{
		Code:       "ERR_NOT_FOUND",
		Message:    "Resource not found",
		StackTrace: &stackTrace,
		Severity:   "HIGH",
	})
	if err != nil {
		t.Fatalf("CreateSystemError: %v", err)
	}

	result, err := bc.ListSystemErrors.Handle(ctx, query.ListSystemErrorsQuery{
		Filter: domain.SystemErrorFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListSystemErrors: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 system error, got %d", result.Total)
	}

	se := result.Errors[0]
	if se.Code != "ERR_NOT_FOUND" {
		t.Errorf("expected code ERR_NOT_FOUND, got %s", se.Code)
	}
	if se.Message != "Resource not found" {
		t.Errorf("expected message 'Resource not found', got %s", se.Message)
	}
	if se.IsResolved {
		t.Errorf("new system error should not be resolved")
	}

	view, err := bc.GetSystemError.Handle(ctx, query.GetSystemErrorQuery{ID: se.ID})
	if err != nil {
		t.Fatalf("GetSystemError: %v", err)
	}
	if view.ID != se.ID {
		t.Errorf("ID mismatch: %s vs %s", view.ID, se.ID)
	}
}

func TestIntegration_ResolveSystemError(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	err := bc.CreateSystemError.Handle(ctx, command.CreateSystemErrorCommand{
		Code:     "ERR_TIMEOUT",
		Message:  "Request timed out",
		Severity: "MEDIUM",
	})
	if err != nil {
		t.Fatalf("CreateSystemError: %v", err)
	}

	list, _ := bc.ListSystemErrors.Handle(ctx, query.ListSystemErrorsQuery{
		Filter: domain.SystemErrorFilter{Limit: 10},
	})
	seID := list.Errors[0].ID

	resolvedBy := uuid.New()
	err = bc.ResolveError.Handle(ctx, command.ResolveErrorCommand{
		ID:         seID,
		ResolvedBy: resolvedBy,
	})
	if err != nil {
		t.Fatalf("ResolveError: %v", err)
	}

	view, _ := bc.GetSystemError.Handle(ctx, query.GetSystemErrorQuery{ID: seID})
	if !view.IsResolved {
		t.Error("system error should be resolved")
	}
}

func TestIntegration_MultipleSystemErrors(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	errors := []command.CreateSystemErrorCommand{
		{Code: "ERR_AUTH", Message: "Unauthorized", Severity: "HIGH"},
		{Code: "ERR_VALIDATION", Message: "Invalid input", Severity: "LOW"},
		{Code: "ERR_INTERNAL", Message: "Internal server error", Severity: "CRITICAL"},
	}

	for _, cmd := range errors {
		if err := bc.CreateSystemError.Handle(ctx, cmd); err != nil {
			t.Fatalf("CreateSystemError (%s): %v", cmd.Code, err)
		}
	}

	result, err := bc.ListSystemErrors.Handle(ctx, query.ListSystemErrorsQuery{
		Filter: domain.SystemErrorFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListSystemErrors: %v", err)
	}
	if result.Total != 3 {
		t.Fatalf("expected 3 system errors, got %d", result.Total)
	}
}
