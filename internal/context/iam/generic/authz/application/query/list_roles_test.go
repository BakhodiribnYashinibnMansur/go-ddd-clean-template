package query

import (
	"context"
	"errors"
	"gct/internal/kernel/infrastructure/logger"
	"testing"

	"gct/internal/context/iam/generic/authz/domain"
	shared "gct/internal/kernel/domain"

	"github.com/stretchr/testify/require"
)

func TestListRolesHandler_WithResults(t *testing.T) {
	t.Parallel()

	desc := "Editor role"
	repo := &mockAuthzReadRepository{
		listRolesFn: func(_ context.Context, _ shared.Pagination) ([]*domain.RoleView, int64, error) {
			return []*domain.RoleView{
				{ID: domain.NewRoleID(), Name: "admin", Description: nil},
				{ID: domain.NewRoleID(), Name: "editor", Description: &desc},
			}, 2, nil
		},
	}

	handler := NewListRolesHandler(repo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListRolesQuery{
		Pagination: shared.Pagination{Limit: 10, Offset: 0},
	})
	require.NoError(t, err)

	if len(result.Roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(result.Roles))
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if result.Roles[0].Name != "admin" {
		t.Errorf("expected first role name 'admin', got '%s'", result.Roles[0].Name)
	}
	if result.Roles[1].Name != "editor" {
		t.Errorf("expected second role name 'editor', got '%s'", result.Roles[1].Name)
	}
	if result.Roles[1].Description == nil || *result.Roles[1].Description != "Editor role" {
		t.Error("expected second role description 'Editor role'")
	}
}

func TestListRolesHandler_Empty(t *testing.T) {
	t.Parallel()

	repo := &mockAuthzReadRepository{
		listRolesFn: func(_ context.Context, _ shared.Pagination) ([]*domain.RoleView, int64, error) {
			return []*domain.RoleView{}, 0, nil
		},
	}

	handler := NewListRolesHandler(repo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListRolesQuery{
		Pagination: shared.Pagination{Limit: 10, Offset: 0},
	})
	require.NoError(t, err)

	if len(result.Roles) != 0 {
		t.Errorf("expected 0 roles, got %d", len(result.Roles))
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
}

func TestListRolesHandler_Pagination(t *testing.T) {
	t.Parallel()

	var capturedPagination shared.Pagination
	repo := &mockAuthzReadRepository{
		listRolesFn: func(_ context.Context, p shared.Pagination) ([]*domain.RoleView, int64, error) {
			capturedPagination = p
			return []*domain.RoleView{
				{ID: domain.NewRoleID(), Name: "viewer"},
			}, 15, nil
		},
	}

	handler := NewListRolesHandler(repo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListRolesQuery{
		Pagination: shared.Pagination{Limit: 5, Offset: 10},
	})
	require.NoError(t, err)

	if capturedPagination.Limit != 5 {
		t.Errorf("expected limit 5, got %d", capturedPagination.Limit)
	}
	if capturedPagination.Offset != 10 {
		t.Errorf("expected offset 10, got %d", capturedPagination.Offset)
	}
	if result.Total != 15 {
		t.Errorf("expected total 15, got %d", result.Total)
	}
	if len(result.Roles) != 1 {
		t.Errorf("expected 1 role in page, got %d", len(result.Roles))
	}
}

func TestListRolesHandler_RepoError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("database timeout")
	repo := &mockAuthzReadRepository{
		listRolesFn: func(_ context.Context, _ shared.Pagination) ([]*domain.RoleView, int64, error) {
			return nil, 0, repoErr
		},
	}

	handler := NewListRolesHandler(repo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListRolesQuery{
		Pagination: shared.Pagination{Limit: 10, Offset: 0},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repo error, got: %v", err)
	}
}
