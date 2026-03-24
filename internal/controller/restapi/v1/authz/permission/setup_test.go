package permission_test

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
// Mock: permission.UseCaseI
// ---------------------------------------------------------------------------

type MockPermUC struct{ mock.Mock }

func (m *MockPermUC) Create(ctx context.Context, p *domain.Permission) error {
	return m.Called(ctx, p).Error(0)
}
func (m *MockPermUC) Get(ctx context.Context, f *domain.PermissionFilter) (*domain.Permission, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Permission), args.Error(1)
}
func (m *MockPermUC) Gets(ctx context.Context, f *domain.PermissionsFilter) ([]*domain.Permission, int, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.Permission), args.Int(1), args.Error(2)
}
func (m *MockPermUC) Update(ctx context.Context, p *domain.Permission) error {
	return m.Called(ctx, p).Error(0)
}
func (m *MockPermUC) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockPermUC) AssignScope(ctx context.Context, permID uuid.UUID, path, method string) error {
	return m.Called(ctx, permID, path, method).Error(0)
}
func (m *MockPermUC) RemoveScope(ctx context.Context, permID uuid.UUID, path, method string) error {
	return m.Called(ctx, permID, path, method).Error(0)
}
func (m *MockPermUC) AssignToRole(ctx context.Context, roleID, permID uuid.UUID) error {
	return m.Called(ctx, roleID, permID).Error(0)
}

// ---------------------------------------------------------------------------
// Mock: authz.UseCaseI
// ---------------------------------------------------------------------------

type MockAuthzUC struct {
	permMock *MockPermUC
}

func (m *MockAuthzUC) Access() accessuc.UseCaseI         { return nil }
func (m *MockAuthzUC) Role() roleuc.UseCaseI             { return nil }
func (m *MockAuthzUC) Permission() permuc.UseCaseI       { return m.permMock }
func (m *MockAuthzUC) Policy() policyuc.UseCaseI         { return nil }
func (m *MockAuthzUC) Relation() relationuc.UseCaseI     { return nil }
func (m *MockAuthzUC) Scope() scopeuc.UseCaseI           { return nil }

func newTestUseCase(permMock *MockPermUC) *usecase.UseCase {
	return &usecase.UseCase{
		Authz: &MockAuthzUC{permMock: permMock},
	}
}
