package query

import (
	"gct/internal/shared/infrastructure/logger"
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/errorcode/domain"

	"github.com/google/uuid"
)

// --- Mocks ---

type mockReadRepo struct {
	view  *domain.ErrorCodeView
	views []*domain.ErrorCodeView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.ErrorCodeView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrErrorCodeNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ domain.ErrorCodeFilter) ([]*domain.ErrorCodeView, int64, error) {
	return m.views, m.total, nil
}

type errorReadRepo struct{ err error }

func (m *errorReadRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.ErrorCodeView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ domain.ErrorCodeFilter) ([]*domain.ErrorCodeView, int64, error) {
	return nil, 0, m.err
}

var errRepo = errors.New("repo failure")

// --- Tests ---

func TestGetErrorCodeHandler_Handle(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &domain.ErrorCodeView{
			ID:         id,
			Code:       "AUTH_001",
			Message:    "unauthorized",
			HTTPStatus: 401,
			Category:   "auth",
			Severity:   "high",
			Retryable:  false,
			RetryAfter: 0,
			Suggestion: "check token",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	handler := NewGetErrorCodeHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetErrorCodeQuery{ID: id})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if result.Code != "AUTH_001" {
		t.Errorf("expected code AUTH_001, got %s", result.Code)
	}
	if result.HTTPStatus != 401 {
		t.Errorf("expected httpStatus 401, got %d", result.HTTPStatus)
	}
	if result.Suggestion != "check token" {
		t.Errorf("expected suggestion 'check token', got %s", result.Suggestion)
	}
}

func TestGetErrorCodeHandler_NotFound(t *testing.T) {
	readRepo := &mockReadRepo{}
	handler := NewGetErrorCodeHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetErrorCodeQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetErrorCodeHandler_RepoError(t *testing.T) {
	readRepo := &errorReadRepo{err: errRepo}
	handler := NewGetErrorCodeHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetErrorCodeQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestGetErrorCodeHandler_AllFieldsMapped(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &domain.ErrorCodeView{
			ID:         id,
			Code:       "DB_001",
			Message:    "connection failed",
			HTTPStatus: 503,
			Category:   "database",
			Severity:   "critical",
			Retryable:  true,
			RetryAfter: 60,
			Suggestion: "wait and retry",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	handler := NewGetErrorCodeHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetErrorCodeQuery{ID: id})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Retryable {
		t.Error("expected retryable true")
	}
	if result.RetryAfter != 60 {
		t.Errorf("expected retryAfter 60, got %d", result.RetryAfter)
	}
	if result.Category != "database" {
		t.Errorf("expected category database, got %s", result.Category)
	}
}
