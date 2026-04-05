package query

import (
	"gct/internal/kernel/infrastructure/logger"
	"context"
	"errors"
	"testing"

	"gct/internal/context/iam/generic/authz/domain"
	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListPermissionsHandler_WithResults(t *testing.T) {
	t.Parallel()

	parentID := uuid.New()
	desc := "Read-only access"
	repo := &mockAuthzReadRepository{
		listPermsFn: func(_ context.Context, _ shared.Pagination) ([]*domain.PermissionView, int64, error) {
			return []*domain.PermissionView{
				{ID: uuid.New(), Name: "users.read", Description: &desc},
				{ID: uuid.New(), ParentID: &parentID, Name: "users.read.self"},
			}, 2, nil
		},
	}

	handler := NewListPermissionsHandler(repo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListPermissionsQuery{
		Pagination: shared.Pagination{Limit: 10, Offset: 0},
	})
	require.NoError(t, err)

	if len(result.Permissions) != 2 {
		t.Fatalf("expected 2 permissions, got %d", len(result.Permissions))
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if result.Permissions[0].Name != "users.read" {
		t.Errorf("expected first perm name 'users.read', got '%s'", result.Permissions[0].Name)
	}
	if result.Permissions[0].Description == nil || *result.Permissions[0].Description != "Read-only access" {
		t.Error("expected first perm description 'Read-only access'")
	}
	if result.Permissions[1].ParentID == nil || *result.Permissions[1].ParentID != parentID {
		t.Error("expected second perm to have correct parent_id")
	}
}

func TestListPermissionsHandler_Empty(t *testing.T) {
	t.Parallel()

	repo := &mockAuthzReadRepository{
		listPermsFn: func(_ context.Context, _ shared.Pagination) ([]*domain.PermissionView, int64, error) {
			return []*domain.PermissionView{}, 0, nil
		},
	}

	handler := NewListPermissionsHandler(repo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListPermissionsQuery{
		Pagination: shared.Pagination{Limit: 10, Offset: 0},
	})
	require.NoError(t, err)

	if len(result.Permissions) != 0 {
		t.Errorf("expected 0 permissions, got %d", len(result.Permissions))
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
}

func TestListPermissionsHandler_RepoError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("query failed")
	repo := &mockAuthzReadRepository{
		listPermsFn: func(_ context.Context, _ shared.Pagination) ([]*domain.PermissionView, int64, error) {
			return nil, 0, repoErr
		},
	}

	handler := NewListPermissionsHandler(repo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListPermissionsQuery{
		Pagination: shared.Pagination{Limit: 10, Offset: 0},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repo error, got: %v", err)
	}
}
