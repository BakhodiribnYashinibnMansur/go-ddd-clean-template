package query

import (
	"gct/internal/kernel/infrastructure/logger"
	"context"
	"errors"
	"testing"

	"gct/internal/context/iam/authz/domain"
	shared "gct/internal/kernel/domain"
	"github.com/stretchr/testify/require"
)

func TestListScopesHandler_WithResults(t *testing.T) {
	t.Parallel()

	repo := &mockAuthzReadRepository{
		listScopesFn: func(_ context.Context, _ shared.Pagination) ([]*domain.ScopeView, int64, error) {
			return []*domain.ScopeView{
				{Path: "/api/v1/users", Method: "GET"},
				{Path: "/api/v1/users", Method: "POST"},
				{Path: "/api/v1/roles", Method: "GET"},
			}, 3, nil
		},
	}

	handler := NewListScopesHandler(repo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListScopesQuery{
		Pagination: shared.Pagination{Limit: 10, Offset: 0},
	})
	require.NoError(t, err)

	if len(result.Scopes) != 3 {
		t.Fatalf("expected 3 scopes, got %d", len(result.Scopes))
	}
	if result.Total != 3 {
		t.Errorf("expected total 3, got %d", result.Total)
	}
	if result.Scopes[0].Path != "/api/v1/users" {
		t.Errorf("expected first scope path '/api/v1/users', got '%s'", result.Scopes[0].Path)
	}
	if result.Scopes[0].Method != "GET" {
		t.Errorf("expected first scope method 'GET', got '%s'", result.Scopes[0].Method)
	}
	if result.Scopes[1].Method != "POST" {
		t.Errorf("expected second scope method 'POST', got '%s'", result.Scopes[1].Method)
	}
}

func TestListScopesHandler_Empty(t *testing.T) {
	t.Parallel()

	repo := &mockAuthzReadRepository{
		listScopesFn: func(_ context.Context, _ shared.Pagination) ([]*domain.ScopeView, int64, error) {
			return []*domain.ScopeView{}, 0, nil
		},
	}

	handler := NewListScopesHandler(repo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListScopesQuery{
		Pagination: shared.Pagination{Limit: 10, Offset: 0},
	})
	require.NoError(t, err)

	if len(result.Scopes) != 0 {
		t.Errorf("expected 0 scopes, got %d", len(result.Scopes))
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
}

func TestListScopesHandler_RepoError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("scope query failed")
	repo := &mockAuthzReadRepository{
		listScopesFn: func(_ context.Context, _ shared.Pagination) ([]*domain.ScopeView, int64, error) {
			return nil, 0, repoErr
		},
	}

	handler := NewListScopesHandler(repo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListScopesQuery{
		Pagination: shared.Pagination{Limit: 10, Offset: 0},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repo error, got: %v", err)
	}
}
