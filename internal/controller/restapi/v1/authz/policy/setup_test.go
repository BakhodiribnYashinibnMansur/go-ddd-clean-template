package policy_test

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
// Mock: policy.UseCaseI
// ---------------------------------------------------------------------------

type MockPolicyUC struct{ mock.Mock }

func (m *MockPolicyUC) Create(ctx context.Context, p *domain.Policy) error {
	return m.Called(ctx, p).Error(0)
}
func (m *MockPolicyUC) Get(ctx context.Context, f *domain.PolicyFilter) (*domain.Policy, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Policy), args.Error(1)
}
func (m *MockPolicyUC) Gets(ctx context.Context, f *domain.PoliciesFilter) ([]*domain.Policy, int, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.Policy), args.Int(1), args.Error(2)
}
func (m *MockPolicyUC) Update(ctx context.Context, p *domain.Policy) error {
	return m.Called(ctx, p).Error(0)
}
func (m *MockPolicyUC) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockPolicyUC) Toggle(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// ---------------------------------------------------------------------------
// Mock: authz.UseCaseI – wires the sub-domain mocks together
// ---------------------------------------------------------------------------

type MockAuthzUC struct {
	policyMock *MockPolicyUC
}

func (m *MockAuthzUC) Access() accessuc.UseCaseI     { return nil }
func (m *MockAuthzUC) Role() roleuc.UseCaseI         { return nil }
func (m *MockAuthzUC) Permission() permuc.UseCaseI   { return nil }
func (m *MockAuthzUC) Policy() policyuc.UseCaseI     { return m.policyMock }
func (m *MockAuthzUC) Relation() relationuc.UseCaseI { return nil }
func (m *MockAuthzUC) Scope() scopeuc.UseCaseI       { return nil }

// newTestUseCase builds a *usecase.UseCase with the given mock policy UC.
func newTestUseCase(policyMock *MockPolicyUC) *usecase.UseCase {
	return &usecase.UseCase{
		Authz: &MockAuthzUC{policyMock: policyMock},
	}
}
