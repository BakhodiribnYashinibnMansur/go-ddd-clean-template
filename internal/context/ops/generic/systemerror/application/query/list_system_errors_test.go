package query

import (
	"gct/internal/kernel/infrastructure/logger"
	"context"
	"testing"
	"time"

	syserrentity "gct/internal/context/ops/generic/systemerror/domain/entity"
	syserrrepo "gct/internal/context/ops/generic/systemerror/domain/repository"

	"github.com/stretchr/testify/require"
)

func TestListSystemErrorsHandler_Handle(t *testing.T) {
	t.Parallel()

	readRepo := &mockReadRepo{
		views: []*syserrrepo.SystemErrorView{
			{ID: syserrentity.NewSystemErrorID(), Code: "ERR_1", Severity: "high", CreatedAt: time.Now()},
			{ID: syserrentity.NewSystemErrorID(), Code: "ERR_2", Severity: "low", CreatedAt: time.Now()},
		},
		total: 2,
	}

	handler := NewListSystemErrorsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListSystemErrorsQuery{
		Filter: syserrrepo.SystemErrorFilter{Limit: 10, Offset: 0},
	})
	require.NoError(t, err)
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if len(result.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(result.Errors))
	}
	if result.Errors[0].Code != "ERR_1" {
		t.Errorf("expected ERR_1, got %s", result.Errors[0].Code)
	}
}

func TestListSystemErrorsHandler_Empty(t *testing.T) {
	t.Parallel()

	readRepo := &mockReadRepo{views: []*syserrrepo.SystemErrorView{}, total: 0}

	handler := NewListSystemErrorsHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListSystemErrorsQuery{
		Filter: syserrrepo.SystemErrorFilter{},
	})
	require.NoError(t, err)
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(result.Errors))
	}
}

func TestListSystemErrorsHandler_WithFilters(t *testing.T) {
	t.Parallel()

	readRepo := &mockReadRepo{
		views: []*syserrrepo.SystemErrorView{
			{ID: syserrentity.NewSystemErrorID(), Code: "ERR_500", Severity: "critical", IsResolved: false, CreatedAt: time.Now()},
		},
		total: 1,
	}

	handler := NewListSystemErrorsHandler(readRepo, logger.Noop())
	code := "ERR_500"
	severity := "critical"
	resolved := false

	result, err := handler.Handle(context.Background(), ListSystemErrorsQuery{
		Filter: syserrrepo.SystemErrorFilter{
			Code:       &code,
			Severity:   &severity,
			IsResolved: &resolved,
			Limit:      10,
		},
	})
	require.NoError(t, err)
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}

func TestListSystemErrorsHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepo}
	handler := NewListSystemErrorsHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListSystemErrorsQuery{Filter: syserrrepo.SystemErrorFilter{}})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
