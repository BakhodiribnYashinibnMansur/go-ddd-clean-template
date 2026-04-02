package query

import (
	"gct/internal/shared/infrastructure/logger"
	"context"
	"testing"
	"time"

	"gct/internal/errorcode/domain"

	"github.com/google/uuid"
)

func TestListErrorCodesHandler_Handle(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*domain.ErrorCodeView{
			{ID: uuid.New(), Code: "ERR_1", Message: "m1", HTTPStatus: 400, Category: "c", Severity: "low", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: uuid.New(), Code: "ERR_2", Message: "m2", HTTPStatus: 500, Category: "c", Severity: "high", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		total: 2,
	}

	handler := NewListErrorCodesHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListErrorCodesQuery{
		Filter: domain.ErrorCodeFilter{Limit: 10, Offset: 0},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if len(result.ErrorCodes) != 2 {
		t.Fatalf("expected 2 error codes, got %d", len(result.ErrorCodes))
	}
	if result.ErrorCodes[0].Code != "ERR_1" {
		t.Errorf("expected ERR_1, got %s", result.ErrorCodes[0].Code)
	}
}

func TestListErrorCodesHandler_Empty(t *testing.T) {
	readRepo := &mockReadRepo{views: []*domain.ErrorCodeView{}, total: 0}

	handler := NewListErrorCodesHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListErrorCodesQuery{
		Filter: domain.ErrorCodeFilter{},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if len(result.ErrorCodes) != 0 {
		t.Errorf("expected 0 error codes, got %d", len(result.ErrorCodes))
	}
}

func TestListErrorCodesHandler_WithFilters(t *testing.T) {
	code := "AUTH_001"
	category := "auth"
	readRepo := &mockReadRepo{
		views: []*domain.ErrorCodeView{
			{ID: uuid.New(), Code: "AUTH_001", Category: "auth", Severity: "high", HTTPStatus: 401, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		total: 1,
	}

	handler := NewListErrorCodesHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListErrorCodesQuery{
		Filter: domain.ErrorCodeFilter{Code: &code, Category: &category, Limit: 10},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}

func TestListErrorCodesHandler_RepoError(t *testing.T) {
	readRepo := &errorReadRepo{err: errRepo}
	handler := NewListErrorCodesHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListErrorCodesQuery{Filter: domain.ErrorCodeFilter{}})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
