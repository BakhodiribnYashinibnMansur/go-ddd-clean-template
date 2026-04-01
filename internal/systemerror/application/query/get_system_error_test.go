package query

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/systemerror/domain"

	"github.com/google/uuid"
)

// --- Mocks ---

type mockReadRepo struct {
	view  *domain.SystemErrorView
	views []*domain.SystemErrorView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.SystemErrorView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrSystemErrorNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ domain.SystemErrorFilter) ([]*domain.SystemErrorView, int64, error) {
	return m.views, m.total, nil
}

type errorReadRepo struct{ err error }

func (m *errorReadRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.SystemErrorView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ domain.SystemErrorFilter) ([]*domain.SystemErrorView, int64, error) {
	return nil, 0, m.err
}

var errRepo = errors.New("repo failure")

// --- Tests: GetSystemError ---

func TestGetSystemErrorHandler_Handle(t *testing.T) {
	id := uuid.New()
	readRepo := &mockReadRepo{
		view: &domain.SystemErrorView{
			ID:       id,
			Code:     "ERR_500",
			Message:  "internal error",
			Severity: "critical",
			CreatedAt: time.Now(),
		},
	}

	handler := NewGetSystemErrorHandler(readRepo)
	result, err := handler.Handle(context.Background(), GetSystemErrorQuery{ID: id})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if result.Code != "ERR_500" {
		t.Errorf("expected code ERR_500, got %s", result.Code)
	}
	if result.Severity != "critical" {
		t.Errorf("expected severity critical, got %s", result.Severity)
	}
}

func TestGetSystemErrorHandler_NotFound(t *testing.T) {
	readRepo := &mockReadRepo{}
	handler := NewGetSystemErrorHandler(readRepo)
	_, err := handler.Handle(context.Background(), GetSystemErrorQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetSystemErrorHandler_RepoError(t *testing.T) {
	readRepo := &errorReadRepo{err: errRepo}
	handler := NewGetSystemErrorHandler(readRepo)
	_, err := handler.Handle(context.Background(), GetSystemErrorQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestGetSystemErrorHandler_AllFieldsMapped(t *testing.T) {
	id := uuid.New()
	stack := "trace"
	svc := "api"
	reqID := uuid.New()
	userID := uuid.New()
	ip := "1.1.1.1"
	path := "/test"
	method := "GET"
	resolvedBy := uuid.New()
	now := time.Now()

	readRepo := &mockReadRepo{
		view: &domain.SystemErrorView{
			ID:          id,
			Code:        "ERR",
			Message:     "msg",
			StackTrace:  &stack,
			Metadata:    map[string]string{"k": "v"},
			Severity:    "high",
			ServiceName: &svc,
			RequestID:   &reqID,
			UserID:      &userID,
			IPAddress:   &ip,
			Path:        &path,
			Method:      &method,
			IsResolved:  true,
			ResolvedAt:  &now,
			ResolvedBy:  &resolvedBy,
			CreatedAt:   now,
		},
	}

	handler := NewGetSystemErrorHandler(readRepo)
	result, err := handler.Handle(context.Background(), GetSystemErrorQuery{ID: id})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StackTrace == nil || *result.StackTrace != "trace" {
		t.Error("stack trace not mapped")
	}
	if result.ServiceName == nil || *result.ServiceName != "api" {
		t.Error("service name not mapped")
	}
	if !result.IsResolved {
		t.Error("expected resolved")
	}
	if result.ResolvedBy == nil || *result.ResolvedBy != resolvedBy {
		t.Error("resolvedBy not mapped")
	}
}
