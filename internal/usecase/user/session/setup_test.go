package session_test

import (
	"context"
	"testing"

	"gct/internal/domain"
	"gct/internal/repo/persistent"
	"gct/internal/repo/persistent/postgres"
	"gct/internal/repo/persistent/postgres/user"
	sessionUC "gct/internal/usecase/user/session"
	"gct/pkg/logger"
	"github.com/stretchr/testify/mock"
)

// MockSessionRepo implements session.RepoI
type MockSessionRepo struct {
	mock.Mock
}

func (m *MockSessionRepo) Create(ctx context.Context, s *domain.Session) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockSessionRepo) Delete(ctx context.Context, f *domain.SessionFilter) error {
	args := m.Called(ctx, f)
	return args.Error(0)
}

func (m *MockSessionRepo) Revoke(ctx context.Context, f *domain.SessionFilter) error {
	args := m.Called(ctx, f)
	return args.Error(0)
}

func (m *MockSessionRepo) Update(ctx context.Context, s *domain.Session) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockSessionRepo) Get(ctx context.Context, f *domain.SessionFilter) (*domain.Session, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *MockSessionRepo) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *MockSessionRepo) Gets(ctx context.Context, f *domain.SessionsFilter) ([]*domain.Session, int, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]*domain.Session), args.Int(1), args.Error(2)
}

func (m *MockSessionRepo) GetByUser(ctx context.Context, userID int64) ([]*domain.Session, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.Session), args.Error(1)
}

func setup(_ *testing.T) (sessionUC.UseCaseI, *MockSessionRepo) {
	sessionRepo := new(MockSessionRepo)

	r := &persistent.Repo{
		Postgres: &postgres.Repo{
			User: &user.User{
				SessionRepo: sessionRepo,
			},
		},
	}

	log := logger.New("debug")

	uc := sessionUC.New(r, log)

	return uc, sessionRepo
}
