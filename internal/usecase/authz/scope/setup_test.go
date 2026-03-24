package scope_test

import (
	"context"
	"testing"

	"gct/internal/domain"
	"gct/internal/repo/persistent"
	"gct/internal/repo/persistent/postgres"
	"gct/internal/repo/persistent/postgres/authz"
	policyRepo "gct/internal/repo/persistent/postgres/authz/policy"
	scopeRepo "gct/internal/repo/persistent/postgres/authz/scope"
	"gct/internal/usecase/authz/scope"
	"gct/internal/shared/infrastructure/logger"

	"github.com/stretchr/testify/mock"
)

// MockScopeRepo implements scopeRepo.RepoI
type MockScopeRepo struct {
	mock.Mock
}

func (m *MockScopeRepo) Create(ctx context.Context, s *domain.Scope) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockScopeRepo) Get(ctx context.Context, filter *domain.ScopeFilter) (*domain.Scope, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Scope), args.Error(1)
}

func (m *MockScopeRepo) Gets(ctx context.Context, filter *domain.ScopesFilter) ([]*domain.Scope, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*domain.Scope), args.Int(1), args.Error(2)
}

func (m *MockScopeRepo) Delete(ctx context.Context, path, method string) error {
	args := m.Called(ctx, path, method)
	return args.Error(0)
}

// Compile-time interface check
var _ scopeRepo.RepoI = (*MockScopeRepo)(nil)

func setup(t *testing.T) (scope.UseCaseI, *MockScopeRepo) {
	t.Helper()

	mockRepo := new(MockScopeRepo)
	log := logger.New("debug")

	repo := &persistent.Repo{
		Postgres: &postgres.Repo{
			Authz: &authz.Authz{
				Scope:  mockRepo,
				Policy: (policyRepo.RepoI)(nil),
			},
		},
	}

	uc := scope.New(repo, log)
	return uc, mockRepo
}
