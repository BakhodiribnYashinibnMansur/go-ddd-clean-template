package client_test

import (
	"errors"
	"testing"

	"gct/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestUseCase_SignIn_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         *domain.SignInIn
		mockUser      *domain.User
		repoError     error
		sessionError  error
		expectError   bool
		validateToken func(t *testing.T, out *domain.SignInOut)
	}{
		{
			name: "success_basic_signin",
			input: &domain.SignInIn{
				Phone:    "123456789",
				Password: "password",
			},
			mockUser: func() *domain.User {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				return &domain.User{
					ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Phone:        stringPtr("123456789"),
					PasswordHash: string(hashedPassword),
				}
			}(),
			expectError: false,
			validateToken: func(t *testing.T, out *domain.SignInOut) {
				t.Helper()
				require.NotEmpty(t, out.AccessToken)
				require.NotEmpty(t, out.RefreshToken)
			},
		},
		{
			name: "success_with_device_id",
			input: &domain.SignInIn{
				Phone:    "123456789",
				Password: "password",
				DeviceID: uuid.MustParse("00000000-0000-0000-0000-000000000123"),
			},
			mockUser: func() *domain.User {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				return &domain.User{
					ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Phone:        stringPtr("123456789"),
					PasswordHash: string(hashedPassword),
				}
			}(),
			expectError: false,
			validateToken: func(t *testing.T, out *domain.SignInOut) {
				t.Helper()
				require.NotEmpty(t, out.AccessToken)
				require.NotEmpty(t, out.RefreshToken)
			},
		},
		{
			name: "success_with_user_agent",
			input: &domain.SignInIn{
				Phone:     "123456789",
				Password:  "password",
				UserAgent: "Mozilla/5.0",
			},
			mockUser: func() *domain.User {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				return &domain.User{
					ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Phone:        stringPtr("123456789"),
					PasswordHash: string(hashedPassword),
				}
			}(),
			expectError: false,
			validateToken: func(t *testing.T, out *domain.SignInOut) {
				t.Helper()
				require.NotEmpty(t, out.AccessToken)
				require.NotEmpty(t, out.RefreshToken)
			},
		},
		{
			name: "success_with_ip",
			input: &domain.SignInIn{
				Phone:    "123456789",
				Password: "password",
				IP:       "192.168.1.1",
			},
			mockUser: func() *domain.User {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				return &domain.User{
					ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Phone:        stringPtr("123456789"),
					PasswordHash: string(hashedPassword),
				}
			}(),
			expectError: false,
			validateToken: func(t *testing.T, out *domain.SignInOut) {
				t.Helper()
				require.NotEmpty(t, out.AccessToken)
				require.NotEmpty(t, out.RefreshToken)
			},
		},
		{
			name: "error_user_not_found",
			input: &domain.SignInIn{
				Phone:    "123456789",
				Password: "password",
			},
			mockUser:    nil,
			repoError:   errors.New("user not found"),
			expectError: true,
		},
		{
			name: "error_invalid_password",
			input: &domain.SignInIn{
				Phone:    "123456789",
				Password: "wrongpassword",
			},
			mockUser: func() *domain.User {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				return &domain.User{
					ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Phone:        stringPtr("123456789"),
					PasswordHash: string(hashedPassword),
				}
			}(),
			expectError: true,
		},
		{
			name: "error_empty_phone",
			input: &domain.SignInIn{
				Phone:    "",
				Password: "password",
			},
			mockUser:    nil,
			repoError:   errors.New("user not found"),
			expectError: true,
		},
		{
			name: "error_empty_password",
			input: &domain.SignInIn{
				Phone:    "123456789",
				Password: "",
			},
			mockUser: func() *domain.User {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				return &domain.User{
					ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Phone:        stringPtr("123456789"),
					PasswordHash: string(hashedPassword),
				}
			}(),
			expectError: true,
		},
		{
			name: "error_session_creation_failed",
			input: &domain.SignInIn{
				Phone:    "123456789",
				Password: "password",
			},
			mockUser: func() *domain.User {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				return &domain.User{
					ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Phone:        stringPtr("123456789"),
					PasswordHash: string(hashedPassword),
				}
			}(),
			sessionError: errors.New("session creation failed"),
			expectError:  true,
		},
		{
			name: "success_all_metadata",
			input: &domain.SignInIn{
				Phone:     "123456789",
				Password:  "password",
				DeviceID:  uuid.MustParse("00000000-0000-0000-0000-000000000123"),
				UserAgent: "Mozilla/5.0",
				IP:        "192.168.1.1",
			},
			mockUser: func() *domain.User {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				return &domain.User{
					ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Phone:        stringPtr("123456789"),
					PasswordHash: string(hashedPassword),
				}
			}(),
			expectError: false,
			validateToken: func(t *testing.T, out *domain.SignInOut) {
				t.Helper()
				require.NotEmpty(t, out.AccessToken)
				require.NotEmpty(t, out.RefreshToken)
			},
		},
	}

	for _, tt := range tests {
		// parallel safety
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			uc, clientRepo, sessionRepo := setup(t)
			ctx := t.Context()

			clientRepo.On("GetByPhone", ctx, tt.input.Phone).Return(tt.mockUser, tt.repoError)

			// Mock session repository if user exists and password is correct
			if tt.mockUser != nil && tt.repoError == nil && tt.input.Password != "wrongpassword" && tt.input.Password != "" {
				if tt.sessionError != nil {
					sessionRepo.On("Create", ctx, mock.AnythingOfType("*domain.Session")).Return(tt.sessionError)
				} else {
					sessionRepo.On("Create", ctx, mock.AnythingOfType("*domain.Session")).Return(nil)
				}
			}

			// act
			out, err := uc.SignIn(ctx, tt.input)

			// assert
			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, out)
			} else {
				require.NoError(t, err)
				require.NotNil(t, out)
				if tt.validateToken != nil {
					tt.validateToken(t, out)
				}
			}

			clientRepo.AssertExpectations(t)
			sessionRepo.AssertExpectations(t)
		})
	}
}
