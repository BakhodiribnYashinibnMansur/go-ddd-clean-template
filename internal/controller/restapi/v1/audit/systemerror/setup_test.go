package systemerror_test

import (
	"context"

	"gct/internal/domain"
	ucsystemerror "gct/internal/usecase/audit/systemerror"

	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// MockSystemErrorUseCase implements ucsystemerror.UseCaseI
// ---------------------------------------------------------------------------

type MockSystemErrorUseCase struct {
	mock.Mock
}

func (m *MockSystemErrorUseCase) Create(ctx context.Context, in *domain.SystemError) error {
	return m.Called(ctx, in).Error(0)
}

func (m *MockSystemErrorUseCase) Gets(ctx context.Context, in *domain.SystemErrorsFilter) ([]*domain.SystemError, int, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*domain.SystemError), args.Get(1).(int), args.Error(2)
}

func (m *MockSystemErrorUseCase) Resolve(ctx context.Context, id string, resolvedBy *string) error {
	return m.Called(ctx, id, resolvedBy).Error(0)
}

// Ensure interface is satisfied at compile time.
var _ ucsystemerror.UseCaseI = (*MockSystemErrorUseCase)(nil)
