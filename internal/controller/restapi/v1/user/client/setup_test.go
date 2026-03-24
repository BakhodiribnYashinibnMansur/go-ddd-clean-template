package client_test

import (
	"context"

	"gct/internal/domain"
	"gct/internal/usecase"
	ucclient "gct/internal/usecase/user/client"
	ucsession "gct/internal/usecase/user/session"
	ucuser "gct/internal/usecase/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// MockClientUseCase implements ucclient.UseCaseI
// ---------------------------------------------------------------------------

type MockClientUseCase struct {
	mock.Mock
}

func (m *MockClientUseCase) Create(ctx context.Context, in *domain.User) error {
	return m.Called(ctx, in).Error(0)
}

func (m *MockClientUseCase) Get(ctx context.Context, in *domain.UserFilter) (*domain.User, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockClientUseCase) Gets(ctx context.Context, in *domain.UsersFilter) ([]*domain.User, int, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*domain.User), args.Get(1).(int), args.Error(2)
}

func (m *MockClientUseCase) Update(ctx context.Context, in *domain.User) error {
	return m.Called(ctx, in).Error(0)
}

func (m *MockClientUseCase) Delete(ctx context.Context, in *domain.UserFilter) error {
	return m.Called(ctx, in).Error(0)
}

func (m *MockClientUseCase) SignIn(ctx context.Context, in *domain.SignInIn) (*domain.SignInOut, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SignInOut), args.Error(1)
}

func (m *MockClientUseCase) SignUp(ctx context.Context, in *domain.SignUpIn) (*domain.SignInOut, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SignInOut), args.Error(1)
}

func (m *MockClientUseCase) SignOut(ctx context.Context, in *domain.SignOutIn) error {
	return m.Called(ctx, in).Error(0)
}

func (m *MockClientUseCase) RotateSession(ctx context.Context, in *domain.RefreshIn) (*domain.SignInOut, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SignInOut), args.Error(1)
}

func (m *MockClientUseCase) GetByPhone(ctx context.Context, in *domain.UserFilter) (*domain.User, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockClientUseCase) ActivateUser(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *MockClientUseCase) SetStatus(ctx context.Context, id uuid.UUID, active bool) error {
	return m.Called(ctx, id, active).Error(0)
}

func (m *MockClientUseCase) BulkAction(ctx context.Context, req domain.BulkActionRequest) error {
	return m.Called(ctx, req).Error(0)
}

func (m *MockClientUseCase) Approve(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockClientUseCase) ChangeRole(ctx context.Context, id, role string) error {
	return m.Called(ctx, id, role).Error(0)
}

// ---------------------------------------------------------------------------
// MockUserUseCase implements ucuser.UseCaseI
// ---------------------------------------------------------------------------

type MockUserUseCase struct {
	client  ucclient.UseCaseI
	session ucsession.UseCaseI
}

func (m *MockUserUseCase) Client() ucclient.UseCaseI   { return m.client }
func (m *MockUserUseCase) Session() ucsession.UseCaseI { return m.session }

// ---------------------------------------------------------------------------
// Helper: build a *usecase.UseCase with mocked User sub-service
// ---------------------------------------------------------------------------

func buildUseCase(clientMock *MockClientUseCase) *usecase.UseCase {
	return &usecase.UseCase{
		User: &MockUserUseCase{client: clientMock},
	}
}

// Ensure interfaces are satisfied at compile time.
var _ ucclient.UseCaseI = (*MockClientUseCase)(nil)
var _ ucuser.UseCaseI = (*MockUserUseCase)(nil)
