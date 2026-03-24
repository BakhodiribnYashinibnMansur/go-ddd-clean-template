package policy_test

import (
	"context"
	"testing"

	"gct/internal/domain"
	"gct/internal/repo/persistent"
	"gct/internal/repo/persistent/postgres"
	"gct/internal/repo/persistent/postgres/authz"
	policyRepo "gct/internal/repo/persistent/postgres/authz/policy"
	scopeRepo "gct/internal/repo/persistent/postgres/authz/scope"
	"gct/internal/usecase/authz/policy"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockPolicyRepo implements policyRepo.RepoI
type MockPolicyRepo struct {
	mock.Mock
}

func (m *MockPolicyRepo) Create(ctx context.Context, p *domain.Policy) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPolicyRepo) Get(ctx context.Context, filter *domain.PolicyFilter) (*domain.Policy, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Policy), args.Error(1)
}

func (m *MockPolicyRepo) Gets(ctx context.Context, filter *domain.PoliciesFilter) ([]*domain.Policy, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*domain.Policy), args.Int(1), args.Error(2)
}

func (m *MockPolicyRepo) Update(ctx context.Context, p *domain.Policy) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPolicyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPolicyRepo) GetByRole(ctx context.Context, roleID uuid.UUID) ([]*domain.Policy, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Policy), args.Error(1)
}

func (m *MockPolicyRepo) Toggle(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Compile-time interface check
var _ policyRepo.RepoI = (*MockPolicyRepo)(nil)

func setup(t *testing.T) (policy.UseCaseI, *MockPolicyRepo) {
	t.Helper()

	mockRepo := new(MockPolicyRepo)
	log := logger.New("debug")

	repo := &persistent.Repo{
		Postgres: &postgres.Repo{
			Authz: &authz.Authz{
				Policy: mockRepo,
				Scope:  (scopeRepo.RepoI)(nil),
			},
		},
	}

	uc := policy.New(repo, log)
	return uc, mockRepo
}
