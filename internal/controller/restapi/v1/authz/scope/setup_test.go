package scope_test

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

	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// Mock: scope.UseCaseI
// ---------------------------------------------------------------------------

type MockScopeUC struct{ mock.Mock }

func (m *MockScopeUC) Create(ctx context.Context, s *domain.Scope) error {
	return m.Called(ctx, s).Error(0)
}
func (m *MockScopeUC) Get(ctx context.Context, f *domain.ScopeFilter) (*domain.Scope, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Scope), args.Error(1)
}
func (m *MockScopeUC) Gets(ctx context.Context, f *domain.ScopesFilter) ([]*domain.Scope, int, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.Scope), args.Int(1), args.Error(2)
}
func (m *MockScopeUC) Delete(ctx context.Context, path, method string) error {
	return m.Called(ctx, path, method).Error(0)
}

// ---------------------------------------------------------------------------
// Mock: authz.UseCaseI – wires the sub-domain mocks together
// ---------------------------------------------------------------------------

type MockAuthzUC struct {
	scopeMock *MockScopeUC
}

func (m *MockAuthzUC) Access() accessuc.UseCaseI         { return nil }
func (m *MockAuthzUC) Role() roleuc.UseCaseI             { return nil }
func (m *MockAuthzUC) Permission() permuc.UseCaseI       { return nil }
func (m *MockAuthzUC) Policy() policyuc.UseCaseI         { return nil }
func (m *MockAuthzUC) Relation() relationuc.UseCaseI     { return nil }
func (m *MockAuthzUC) Scope() scopeuc.UseCaseI           { return m.scopeMock }

// newTestUseCase builds a *usecase.UseCase with the given mock scope UC.
func newTestUseCase(scopeMock *MockScopeUC) *usecase.UseCase {
	return &usecase.UseCase{
		Authz: &MockAuthzUC{scopeMock: scopeMock},
	}
}
