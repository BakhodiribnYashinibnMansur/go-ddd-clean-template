package relation_test

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
// Mock: relation.UseCaseI
// ---------------------------------------------------------------------------

type MockRelationUC struct{ mock.Mock }

func (m *MockRelationUC) Create(ctx context.Context, r *domain.Relation) error {
	return m.Called(ctx, r).Error(0)
}
func (m *MockRelationUC) Get(ctx context.Context, f *domain.RelationFilter) (*domain.Relation, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Relation), args.Error(1)
}
func (m *MockRelationUC) Gets(ctx context.Context, f *domain.RelationsFilter) ([]*domain.Relation, int, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.Relation), args.Int(1), args.Error(2)
}
func (m *MockRelationUC) Update(ctx context.Context, r *domain.Relation) error {
	return m.Called(ctx, r).Error(0)
}
func (m *MockRelationUC) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockRelationUC) AddUser(ctx context.Context, userID, relationID uuid.UUID) error {
	return m.Called(ctx, userID, relationID).Error(0)
}
func (m *MockRelationUC) RemoveUser(ctx context.Context, userID, relationID uuid.UUID) error {
	return m.Called(ctx, userID, relationID).Error(0)
}

// ---------------------------------------------------------------------------
// Mock: authz.UseCaseI – wires the sub-domain mocks together
// ---------------------------------------------------------------------------

type MockAuthzUC struct {
	relationMock *MockRelationUC
}

func (m *MockAuthzUC) Access() accessuc.UseCaseI         { return nil }
func (m *MockAuthzUC) Role() roleuc.UseCaseI             { return nil }
func (m *MockAuthzUC) Permission() permuc.UseCaseI       { return nil }
func (m *MockAuthzUC) Policy() policyuc.UseCaseI         { return nil }
func (m *MockAuthzUC) Relation() relationuc.UseCaseI     { return m.relationMock }
func (m *MockAuthzUC) Scope() scopeuc.UseCaseI           { return nil }

// newTestUseCase builds a *usecase.UseCase with the given mock relation UC.
func newTestUseCase(relationMock *MockRelationUC) *usecase.UseCase {
	return &usecase.UseCase{
		Authz: &MockAuthzUC{relationMock: relationMock},
	}
}
