package query

import (
	"context"
	"errors"
	"gct/internal/kernel/infrastructure/logger"
	"testing"

	authzentity "gct/internal/context/iam/generic/authz/domain/entity"
	authzrepo "gct/internal/context/iam/generic/authz/domain/repository"
	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock AuthzReadRepository (inline, scoped to query tests)
// ---------------------------------------------------------------------------

type mockAuthzReadRepository struct {
	getRoleFn               func(ctx context.Context, id authzentity.RoleID) (*authzrepo.RoleView, error)
	listRolesFn             func(ctx context.Context, p shared.Pagination) ([]*authzrepo.RoleView, int64, error)
	getPermissionFn         func(ctx context.Context, id authzentity.PermissionID) (*authzrepo.PermissionView, error)
	listPermsFn             func(ctx context.Context, p shared.Pagination) ([]*authzrepo.PermissionView, int64, error)
	listPoliciesFn          func(ctx context.Context, p shared.Pagination) ([]*authzrepo.PolicyView, int64, error)
	listScopesFn            func(ctx context.Context, p shared.Pagination) ([]*authzrepo.ScopeView, int64, error)
	checkAccessFn           func(ctx context.Context, roleID authzentity.RoleID, path, method string, evalCtx authzentity.EvaluationContext) (bool, error)
	findPoliciesByPermIDsFn func(ctx context.Context, permissionIDs []authzentity.PermissionID) ([]*authzentity.Policy, error)
}

func (m *mockAuthzReadRepository) GetRole(ctx context.Context, id authzentity.RoleID) (*authzrepo.RoleView, error) {
	if m.getRoleFn != nil {
		return m.getRoleFn(ctx, id)
	}
	return nil, authzentity.ErrRoleNotFound
}

func (m *mockAuthzReadRepository) ListRoles(ctx context.Context, p shared.Pagination) ([]*authzrepo.RoleView, int64, error) {
	if m.listRolesFn != nil {
		return m.listRolesFn(ctx, p)
	}
	return nil, 0, nil
}

func (m *mockAuthzReadRepository) GetPermission(ctx context.Context, id authzentity.PermissionID) (*authzrepo.PermissionView, error) {
	if m.getPermissionFn != nil {
		return m.getPermissionFn(ctx, id)
	}
	return nil, authzentity.ErrPermissionNotFound
}

func (m *mockAuthzReadRepository) ListPermissions(ctx context.Context, p shared.Pagination) ([]*authzrepo.PermissionView, int64, error) {
	if m.listPermsFn != nil {
		return m.listPermsFn(ctx, p)
	}
	return nil, 0, nil
}

func (m *mockAuthzReadRepository) ListPolicies(ctx context.Context, p shared.Pagination) ([]*authzrepo.PolicyView, int64, error) {
	if m.listPoliciesFn != nil {
		return m.listPoliciesFn(ctx, p)
	}
	return nil, 0, nil
}

func (m *mockAuthzReadRepository) ListScopes(ctx context.Context, p shared.Pagination) ([]*authzrepo.ScopeView, int64, error) {
	if m.listScopesFn != nil {
		return m.listScopesFn(ctx, p)
	}
	return nil, 0, nil
}

func (m *mockAuthzReadRepository) CheckAccess(ctx context.Context, roleID authzentity.RoleID, path, method string, evalCtx authzentity.EvaluationContext) (bool, error) {
	if m.checkAccessFn != nil {
		return m.checkAccessFn(ctx, roleID, path, method, evalCtx)
	}
	return false, nil
}

func (m *mockAuthzReadRepository) FindPoliciesByPermissionIDs(ctx context.Context, permissionIDs []authzentity.PermissionID) ([]*authzentity.Policy, error) {
	if m.findPoliciesByPermIDsFn != nil {
		return m.findPoliciesByPermIDsFn(ctx, permissionIDs)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// Tests: GetRoleHandler
// ---------------------------------------------------------------------------

func TestGetRoleHandler_Found(t *testing.T) {
	t.Parallel()

	roleID := authzentity.NewRoleID()
	desc := "Admin role"
	repo := &mockAuthzReadRepository{
		getRoleFn: func(_ context.Context, id authzentity.RoleID) (*authzrepo.RoleView, error) {
			if id == roleID {
				return &authzrepo.RoleView{
					ID:          roleID,
					Name:        "admin",
					Description: &desc,
				}, nil
			}
			return nil, authzentity.ErrRoleNotFound
		},
	}

	handler := NewGetRoleHandler(repo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetRoleQuery{ID: authzentity.RoleID(roleID)})
	require.NoError(t, err)

	if result.ID != uuid.UUID(roleID) {
		t.Errorf("expected ID %s, got %s", roleID, result.ID)
	}
	if result.Name != "admin" {
		t.Errorf("expected name 'admin', got '%s'", result.Name)
	}
	if result.Description == nil || *result.Description != "Admin role" {
		t.Error("expected description 'Admin role'")
	}
}

func TestGetRoleHandler_NotFound(t *testing.T) {
	t.Parallel()

	repo := &mockAuthzReadRepository{
		getRoleFn: func(_ context.Context, _ authzentity.RoleID) (*authzrepo.RoleView, error) {
			return nil, authzentity.ErrRoleNotFound
		},
	}

	handler := NewGetRoleHandler(repo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetRoleQuery{ID: authzentity.RoleID(uuid.New())})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, authzentity.ErrRoleNotFound) {
		t.Errorf("expected ErrRoleNotFound, got: %v", err)
	}
}

func TestGetRoleHandler_RepoError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("database connection failed")
	repo := &mockAuthzReadRepository{
		getRoleFn: func(_ context.Context, _ authzentity.RoleID) (*authzrepo.RoleView, error) {
			return nil, repoErr
		},
	}

	handler := NewGetRoleHandler(repo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetRoleQuery{ID: authzentity.RoleID(uuid.New())})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repo error, got: %v", err)
	}
}
