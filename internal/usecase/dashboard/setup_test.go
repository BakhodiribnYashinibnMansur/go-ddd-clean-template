package dashboard_test

import (
	"context"
	"testing"

	"gct/internal/domain"
	"gct/internal/usecase/dashboard"
	"gct/internal/shared/infrastructure/logger"

	"github.com/stretchr/testify/mock"
)

// MockRepo implements dashboard.Repository
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Get(ctx context.Context) (domain.DashboardStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(domain.DashboardStats), args.Error(1)
}

func setup(t *testing.T) (dashboard.UseCaseI, *MockRepo) {
	t.Helper()
	repo := new(MockRepo)
	log := logger.New("debug")
	uc := dashboard.New(repo, log)
	return uc, repo
}
