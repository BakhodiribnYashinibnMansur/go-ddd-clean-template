package query

import (
	"context"
	"errors"
	"gct/internal/kernel/infrastructure/logger"
	"testing"
	"time"

	errcodeentity "gct/internal/context/admin/supporting/errorcode/domain/entity"
	errcoderepo "gct/internal/context/admin/supporting/errorcode/domain/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockReadRepo struct {
	view  *errcoderepo.ErrorCodeView
	views []*errcoderepo.ErrorCodeView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id errcodeentity.ErrorCodeID) (*errcoderepo.ErrorCodeView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, errcodeentity.ErrErrorCodeNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ errcoderepo.ErrorCodeFilter) ([]*errcoderepo.ErrorCodeView, int64, error) {
	return m.views, m.total, nil
}

type errorReadRepo struct{ err error }

func (m *errorReadRepo) FindByID(_ context.Context, _ errcodeentity.ErrorCodeID) (*errcoderepo.ErrorCodeView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ errcoderepo.ErrorCodeFilter) ([]*errcoderepo.ErrorCodeView, int64, error) {
	return nil, 0, m.err
}

var errRepo = errors.New("repo failure")

// --- Tests ---

func TestGetErrorCodeHandler_Handle(t *testing.T) {
	t.Parallel()

	id := errcodeentity.NewErrorCodeID()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &errcoderepo.ErrorCodeView{
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
	result, err := handler.Handle(context.Background(), GetErrorCodeQuery{ID: errcodeentity.ErrorCodeID(id)})
	require.NoError(t, err)
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
	t.Parallel()

	readRepo := &mockReadRepo{}
	handler := NewGetErrorCodeHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetErrorCodeQuery{ID: errcodeentity.ErrorCodeID(uuid.New())})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetErrorCodeHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepo}
	handler := NewGetErrorCodeHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetErrorCodeQuery{ID: errcodeentity.ErrorCodeID(uuid.New())})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestGetErrorCodeHandler_AllFieldsMapped(t *testing.T) {
	t.Parallel()

	id := errcodeentity.NewErrorCodeID()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &errcoderepo.ErrorCodeView{
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
	result, err := handler.Handle(context.Background(), GetErrorCodeQuery{ID: errcodeentity.ErrorCodeID(id)})
	require.NoError(t, err)
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
