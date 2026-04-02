package query

import (
	"gct/internal/shared/infrastructure/logger"
	"context"
	"errors"
	"testing"

	"gct/internal/authz/domain"
	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Mock AuthzReadRepository (inline, scoped to query tests)
// ---------------------------------------------------------------------------

type mockAuthzReadRepository struct {
	getRoleFn        func(ctx context.Context, id uuid.UUID) (*domain.RoleView, error)
	listRolesFn      func(ctx context.Context, p shared.Pagination) ([]*domain.RoleView, int64, error)
	getPermissionFn  func(ctx context.Context, id uuid.UUID) (*domain.PermissionView, error)
	listPermsFn      func(ctx context.Context, p shared.Pagination) ([]*domain.PermissionView, int64, error)
	listPoliciesFn   func(ctx context.Context, p shared.Pagination) ([]*domain.PolicyView, int64, error)
	listScopesFn     func(ctx context.Context, p shared.Pagination) ([]*domain.ScopeView, int64, error)
	checkAccessFn              func(ctx context.Context, roleID uuid.UUID, path, method string, evalCtx domain.EvaluationContext) (bool, error)
	findPoliciesByPermIDsFn    func(ctx context.Context, permissionIDs []uuid.UUID) ([]*domain.Policy, error)
}

func (m *mockAuthzReadRepository) GetRole(ctx context.Context, id uuid.UUID) (*domain.RoleView, error) {
	if m.getRoleFn != nil {
		return m.getRoleFn(ctx, id)
	}
	return nil, domain.ErrRoleNotFound
}

func (m *mockAuthzReadRepository) ListRoles(ctx context.Context, p shared.Pagination) ([]*domain.RoleView, int64, error) {
	if m.listRolesFn != nil {
		return m.listRolesFn(ctx, p)
	}
	return nil, 0, nil
}

func (m *mockAuthzReadRepository) GetPermission(ctx context.Context, id uuid.UUID) (*domain.PermissionView, error) {
	if m.getPermissionFn != nil {
		return m.getPermissionFn(ctx, id)
	}
	return nil, domain.ErrPermissionNotFound
}

func (m *mockAuthzReadRepository) ListPermissions(ctx context.Context, p shared.Pagination) ([]*domain.PermissionView, int64, error) {
	if m.listPermsFn != nil {
		return m.listPermsFn(ctx, p)
	}
	return nil, 0, nil
}

func (m *mockAuthzReadRepository) ListPolicies(ctx context.Context, p shared.Pagination) ([]*domain.PolicyView, int64, error) {
	if m.listPoliciesFn != nil {
		return m.listPoliciesFn(ctx, p)
	}
	return nil, 0, nil
}

func (m *mockAuthzReadRepository) ListScopes(ctx context.Context, p shared.Pagination) ([]*domain.ScopeView, int64, error) {
	if m.listScopesFn != nil {
		return m.listScopesFn(ctx, p)
	}
	return nil, 0, nil
}

func (m *mockAuthzReadRepository) CheckAccess(ctx context.Context, roleID uuid.UUID, path, method string, evalCtx domain.EvaluationContext) (bool, error) {
	if m.checkAccessFn != nil {
		return m.checkAccessFn(ctx, roleID, path, method, evalCtx)
	}
	return false, nil
}

func (m *mockAuthzReadRepository) FindPoliciesByPermissionIDs(ctx context.Context, permissionIDs []uuid.UUID) ([]*domain.Policy, error) {
	if m.findPoliciesByPermIDsFn != nil {
		return m.findPoliciesByPermIDsFn(ctx, permissionIDs)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// Tests: GetRoleHandler
// ---------------------------------------------------------------------------

func TestGetRoleHandler_Found(t *testing.T) {
	roleID := uuid.New()
	desc := "Admin role"
	repo := &mockAuthzReadRepository{
		getRoleFn: func(_ context.Context, id uuid.UUID) (*domain.RoleView, error) {
			if id == roleID {
				return &domain.RoleView{
					ID:          roleID,
					Name:        "admin",
					Description: &desc,
				}, nil
			}
			return nil, domain.ErrRoleNotFound
		},
	}

	handler := NewGetRoleHandler(repo, logger.Noop())
	result, err := handler.Handle(context.Background(), GetRoleQuery{ID: roleID})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result.ID != roleID {
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
	repo := &mockAuthzReadRepository{
		getRoleFn: func(_ context.Context, _ uuid.UUID) (*domain.RoleView, error) {
			return nil, domain.ErrRoleNotFound
		},
	}

	handler := NewGetRoleHandler(repo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetRoleQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrRoleNotFound) {
		t.Errorf("expected ErrRoleNotFound, got: %v", err)
	}
}

func TestGetRoleHandler_RepoError(t *testing.T) {
	repoErr := errors.New("database connection failed")
	repo := &mockAuthzReadRepository{
		getRoleFn: func(_ context.Context, _ uuid.UUID) (*domain.RoleView, error) {
			return nil, repoErr
		},
	}

	handler := NewGetRoleHandler(repo, logger.Noop())
	_, err := handler.Handle(context.Background(), GetRoleQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("expected repo error, got: %v", err)
	}
}
