package query

import (
	"gct/internal/kernel/infrastructure/logger"
	"context"
	"errors"
	"testing"
	"time"

	syserrentity "gct/internal/context/ops/generic/systemerror/domain/entity"
	syserrrepo "gct/internal/context/ops/generic/systemerror/domain/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockReadRepo struct {
	view  *syserrrepo.SystemErrorView
	views []*syserrrepo.SystemErrorView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id syserrentity.SystemErrorID) (*syserrrepo.SystemErrorView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, syserrentity.ErrSystemErrorNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ syserrrepo.SystemErrorFilter) ([]*syserrrepo.SystemErrorView, int64, error) {
	return m.views, m.total, nil
}

type errorReadRepo struct{ err error }

func (m *errorReadRepo) FindByID(_ context.Context, _ syserrentity.SystemErrorID) (*syserrrepo.SystemErrorView, error) {
	return nil, m.err
}

func (m *errorReadRepo) List(_ context.Context, _ syserrrepo.SystemErrorFilter) ([]*syserrrepo.SystemErrorView, int64, error) {
	return nil, 0, m.err
}

var errRepo = errors.New("repo failure")

// --- Tests: GetSystemError ---

func TestGetSystemErrorHandler_Handle(t *testing.T) {
	t.Parallel()

	id := syserrentity.NewSystemErrorID()
	readRepo := &mockReadRepo{
		view: &syserrrepo.SystemErrorView{
			ID:       id,
			Code:     "ERR_500",
			Message:  "internal error",
			Severity: "critical",
			CreatedAt: time.Now(),
		},
	}

	handler := NewGetSystemErrorHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetSystemErrorQuery{ID: id})
	require.NoError(t, err)
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
	t.Parallel()

	readRepo := &mockReadRepo{}
	handler := NewGetSystemErrorHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetSystemErrorQuery{ID: syserrentity.NewSystemErrorID()})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestGetSystemErrorHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepo}
	handler := NewGetSystemErrorHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetSystemErrorQuery{ID: syserrentity.NewSystemErrorID()})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}

func TestGetSystemErrorHandler_AllFieldsMapped(t *testing.T) {
	t.Parallel()

	id := syserrentity.NewSystemErrorID()
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
		view: &syserrrepo.SystemErrorView{
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

	handler := NewGetSystemErrorHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetSystemErrorQuery{ID: id})
	require.NoError(t, err)
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
