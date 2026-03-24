package role_test

import (
	"context"

	"gct/internal/domain"
	"gct/internal/usecase"
	accessuc "gct/internal/usecase/authz/access"
	permuc "gct/internal/usecase/authz/permission"
	policyuc "gct/internal/usecase/authz/policy"
	relationuc "gct/internal/usecase/authz/relation"
	roleuc "gct/internal/usecase/authz/role"
	scopeuc "gct/internal/usecase/authz/scope"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// Mock: role.UseCaseI
// ---------------------------------------------------------------------------

type MockRoleUC struct{ mock.Mock }

func (m *MockRoleUC) Create(ctx context.Context, r *domain.Role) error {
	return m.Called(ctx, r).Error(0)
}
func (m *MockRoleUC) Get(ctx context.Context, f *domain.RoleFilter) (*domain.Role, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}
func (m *MockRoleUC) Gets(ctx context.Context, f *domain.RolesFilter) ([]*domain.Role, int, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.Role), args.Int(1), args.Error(2)
}
func (m *MockRoleUC) Update(ctx context.Context, r *domain.Role) error {
	return m.Called(ctx, r).Error(0)
}
func (m *MockRoleUC) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockRoleUC) Assign(ctx context.Context, userID, roleID uuid.UUID) error {
	return m.Called(ctx, userID, roleID).Error(0)
}
func (m *MockRoleUC) AddPermission(ctx context.Context, roleID, permID uuid.UUID) error {
	return m.Called(ctx, roleID, permID).Error(0)
}
func (m *MockRoleUC) RemovePermission(ctx context.Context, roleID, permID uuid.UUID) error {
	return m.Called(ctx, roleID, permID).Error(0)
}

// ---------------------------------------------------------------------------
// Mock: authz.UseCaseI – wires the sub-domain mocks together
// ---------------------------------------------------------------------------

type MockAuthzUC struct {
	roleMock *MockRoleUC
}

func (m *MockAuthzUC) Access() accessuc.UseCaseI         { return nil }
func (m *MockAuthzUC) Role() roleuc.UseCaseI             { return m.roleMock }
func (m *MockAuthzUC) Permission() permuc.UseCaseI       { return nil }
func (m *MockAuthzUC) Policy() policyuc.UseCaseI         { return nil }
func (m *MockAuthzUC) Relation() relationuc.UseCaseI     { return nil }
func (m *MockAuthzUC) Scope() scopeuc.UseCaseI           { return nil }

// newTestUseCase builds a *usecase.UseCase with the given mock role UC.
func newTestUseCase(roleMock *MockRoleUC) *usecase.UseCase {
	return &usecase.UseCase{
		Authz: &MockAuthzUC{roleMock: roleMock},
	}
}
