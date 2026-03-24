package dashboard_test

import (
	"context"

	"gct/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockUseCase implements ucdashboard.UseCaseI for testing.
type MockUseCase struct {
	mock.Mock
}

func (m *MockUseCase) Get(ctx context.Context) (domain.DashboardStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(domain.DashboardStats), args.Error(1)
}
