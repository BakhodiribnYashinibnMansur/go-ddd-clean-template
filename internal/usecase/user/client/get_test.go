package client_test

import (
	"errors"
	"testing"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Get_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		filter       *domain.UserFilter
		mockUser     *domain.User
		repoError    error
		expectError  bool
		validateUser func(t *testing.T, got *domain.User)
	}{
		{
			name: "success_valid_id",
			filter: &domain.UserFilter{
				ID: uuidPtr(uuid.MustParse("00000000-0000-0000-0000-000000000001")),
			},
			mockUser:    &domain.User{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Phone: stringPtr("123456789")},
			repoError:   nil,
			expectError: false,
			validateUser: func(t *testing.T, got *domain.User) {
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000001"), got.ID)
				require.Equal(t, "123456789", *got.Phone)
			},
		},
		{
			name: "success_zero_id",
			filter: &domain.UserFilter{
				ID: uuidPtr(uuid.UUID{}), // zero UUID
			},
			mockUser:    &domain.User{ID: uuid.UUID{}, Phone: stringPtr("987654321")},
			repoError:   nil,
			expectError: false,
			validateUser: func(t *testing.T, got *domain.User) {
				require.Equal(t, uuid.UUID{}, got.ID)
				require.Equal(t, "987654321", *got.Phone)
			},
		},
		{
			name: "success_negative_id",
			filter: &domain.UserFilter{
				ID: uuidPtr(uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")), // max UUID equivalent
			},
			mockUser:    &domain.User{ID: uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"), Phone: stringPtr("555555555")},
			repoError:   nil,
			expectError: false,
			validateUser: func(t *testing.T, got *domain.User) {
				require.Equal(t, uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"), got.ID)
				require.Equal(t, "555555555", *got.Phone)
			},
		},
		{
			name: "success_nil_id_pointer",
			filter: &domain.UserFilter{
				ID: nil,
			},
			mockUser:    &domain.User{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Phone: stringPtr("111111111")},
			repoError:   nil,
			expectError: false,
			validateUser: func(t *testing.T, got *domain.User) {
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000001"), got.ID)
				require.Equal(t, "111111111", *got.Phone)
			},
		},
		{
			name: "error_repo_returns_error",
			filter: &domain.UserFilter{
				ID: uuidPtr(uuid.MustParse("00000000-0000-0000-0000-000000000001")),
			},
			mockUser:    nil,
			repoError:   errors.New("user not found"),
			expectError: true,
			validateUser: func(t *testing.T, got *domain.User) {
				// No validation expected on error
			},
		},
		{
			name: "success_nil_user_from_repo",
			filter: &domain.UserFilter{
				ID: uuidPtr(uuid.MustParse("00000000-0000-0000-0000-000000000999")), // valid UUID
			},
			mockUser:    nil,
			repoError:   nil,
			expectError: false,
			validateUser: func(t *testing.T, got *domain.User) {
				require.Nil(t, got)
			},
		},
		{
			name: "success_user_with_username",
			filter: &domain.UserFilter{
				ID: uuidPtr(uuid.MustParse("00000000-0000-0000-0000-000000000002")),
			},
			mockUser: &domain.User{
				ID:       uuid.MustParse("00000000-0000-0000-0000-000000000002"),
				Phone:    stringPtr("222222222"),
				Username: stringPtr("testuser"),
			},
			repoError:   nil,
			expectError: false,
			validateUser: func(t *testing.T, got *domain.User) {
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000002"), got.ID)
				require.Equal(t, "222222222", *got.Phone)
				require.NotNil(t, got.Username)
				require.Equal(t, "testuser", *got.Username)
			},
		},
		{
			name: "success_user_with_nil_username",
			filter: &domain.UserFilter{
				ID: uuidPtr(uuid.MustParse("00000000-0000-0000-0000-000000000003")),
			},
			mockUser: &domain.User{
				ID:       uuid.MustParse("00000000-0000-0000-0000-000000000003"),
				Phone:    stringPtr("333333333"),
				Username: nil,
			},
			repoError:   nil,
			expectError: false,
			validateUser: func(t *testing.T, got *domain.User) {
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000003"), got.ID)
				require.Equal(t, "333333333", *got.Phone)
				require.Nil(t, got.Username)
			},
		},
	}

	for _, tt := range tests {
		tt := tt // parallel safety
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			uc, clientRepo, _ := setup(t)
			ctx := t.Context()

			clientRepo.On("Get", ctx, tt.filter).Return(tt.mockUser, tt.repoError).Once()

			// act
			got, err := uc.Get(ctx, tt.filter)

			// assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validateUser != nil {
					tt.validateUser(t, got)
				}
			}

			clientRepo.AssertExpectations(t)
		})
	}
}
