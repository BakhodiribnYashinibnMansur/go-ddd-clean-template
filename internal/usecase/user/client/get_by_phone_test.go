package client_test

import (
	"errors"
	"testing"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUseCase_GetByPhone_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		phone       string
		mockUser    *domain.User
		repoError   error
		expectError bool
		validateOut func(t *testing.T, u *domain.User)
	}{
		{
			name:  "success_basic_get",
			phone: "123456789",
			mockUser: &domain.User{
				ID:    uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Phone: stringPtr("123456789"),
			},
			repoError:   nil,
			expectError: false,
			validateOut: func(t *testing.T, u *domain.User) {
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000001"), u.ID)
				require.Equal(t, "123456789", *u.Phone)
			},
		},
		{
			name:  "success_with_username",
			phone: "123456789",
			mockUser: &domain.User{
				ID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Phone:    stringPtr("123456789"),
				Username: stringPtr("testuser"),
			},
			repoError:   nil,
			expectError: false,
			validateOut: func(t *testing.T, u *domain.User) {
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000001"), u.ID)
				require.Equal(t, "123456789", *u.Phone)
				require.NotNil(t, u.Username)
				require.Equal(t, "testuser", *u.Username)
			},
		},
		{
			name:        "error_user_not_found",
			phone:       "123456789",
			mockUser:    nil,
			repoError:   errors.New("user not found"),
			expectError: true,
		},
		{
			name:        "error_empty_phone",
			phone:       "",
			mockUser:    nil,
			repoError:   nil, // No repo call made due to validation
			expectError: true,
		},
		{
			name:  "success_long_phone",
			phone: "98765432109876543210",
			mockUser: &domain.User{
				ID:    uuid.MustParse("00000000-0000-0000-0000-000000000002"),
				Phone: stringPtr("98765432109876543210"),
			},
			repoError:   nil,
			expectError: false,
			validateOut: func(t *testing.T, u *domain.User) {
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000002"), u.ID)
				require.Equal(t, "98765432109876543210", *u.Phone)
			},
		},
		{
			name:  "success_with_password_hash",
			phone: "123456789",
			mockUser: &domain.User{
				ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Phone:        stringPtr("123456789"),
				PasswordHash: "$2a$10$hashedpassword",
			},
			repoError:   nil,
			expectError: false,
			validateOut: func(t *testing.T, u *domain.User) {
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000001"), u.ID)
				require.Equal(t, "123456789", *u.Phone)
				require.NotEmpty(t, u.PasswordHash)
			},
		},
		{
			name:        "error_database_connection",
			phone:       "123456789",
			mockUser:    nil,
			repoError:   errors.New("database connection failed"),
			expectError: true,
		},
		{
			name:  "success_numeric_phone_only",
			phone: "9876543210",
			mockUser: &domain.User{
				ID:    uuid.MustParse("00000000-0000-0000-0000-000000000003"),
				Phone: stringPtr("9876543210"),
			},
			repoError:   nil,
			expectError: false,
			validateOut: func(t *testing.T, u *domain.User) {
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000003"), u.ID)
				require.Equal(t, "9876543210", *u.Phone)
			},
		},
		{
			name:  "success_with_all_fields",
			phone: "123456789",
			mockUser: &domain.User{
				ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Username:     stringPtr("fulluser"),
				Phone:        stringPtr("123456789"),
				PasswordHash: "$2a$10$hashedpassword",
				Salt:         stringPtr("salt"),
			},
			repoError:   nil,
			expectError: false,
			validateOut: func(t *testing.T, u *domain.User) {
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000001"), u.ID)
				require.NotNil(t, u.Username)
				require.Equal(t, "fulluser", *u.Username)
				require.Equal(t, "123456789", *u.Phone)
				require.NotEmpty(t, u.PasswordHash)
				require.NotNil(t, u.Salt)
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
			filter := &domain.UserFilter{Phone: &tt.phone}

			// Only mock repo call if phone is not empty (validation passes)
			if tt.phone != "" {
				clientRepo.On("GetByPhone", ctx, tt.phone).Return(tt.mockUser, tt.repoError)
			}

			// act
			got, err := uc.GetByPhone(ctx, filter)

			// assert
			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				if tt.validateOut != nil {
					tt.validateOut(t, got)
				}
			}

			// Only assert expectations if we set up any mocks
			if tt.phone != "" {
				clientRepo.AssertExpectations(t)
			}
		})
	}
}
