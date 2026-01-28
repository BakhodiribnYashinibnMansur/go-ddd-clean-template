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
		skipRepoMock  bool
		validateToken func(t *testing.T, out *domain.SignInOut)
	}{
		{
			name: "success_basic_signin",
			input: &domain.SignInIn{
				Login:    stringPtr("123456789"),
				Password: stringPtr("password"),
			},
			mockUser: func() *domain.User {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				return &domain.User{
					ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Phone:        stringPtr("123456789"),
					PasswordHash: string(hashedPassword),
					IsApproved:   true,
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
			input: func() *domain.SignInIn {
				in := &domain.SignInIn{
					Login:    stringPtr("123456789"),
					Password: stringPtr("password"),
				}
				in.Session.DeviceID = uuid.MustParse("00000000-0000-0000-0000-000000000123")
				return in
			}(),
			mockUser: func() *domain.User {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				return &domain.User{
					ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Phone:        stringPtr("123456789"),
					PasswordHash: string(hashedPassword),
					IsApproved:   true,
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
			input: func() *domain.SignInIn {
				in := &domain.SignInIn{
					Login:    stringPtr("123456789"),
					Password: stringPtr("password"),
				}
				in.Session.UserAgent = "Mozilla/5.0"
				return in
			}(),
			mockUser: func() *domain.User {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				return &domain.User{
					ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Phone:        stringPtr("123456789"),
					PasswordHash: string(hashedPassword),
					IsApproved:   true,
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
			input: func() *domain.SignInIn {
				in := &domain.SignInIn{
					Login:    stringPtr("123456789"),
					Password: stringPtr("password"),
				}
				in.Session.IP = "192.168.1.1"
				return in
			}(),
			mockUser: func() *domain.User {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				return &domain.User{
					ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Phone:        stringPtr("123456789"),
					PasswordHash: string(hashedPassword),
					IsApproved:   true,
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
				Login:    stringPtr("123456789"),
				Password: stringPtr("password"),
			},
			mockUser:    nil,
			repoError:   errors.New("user not found"),
			expectError: true,
		},
		{
			name: "error_invalid_password",
			input: &domain.SignInIn{
				Login:    stringPtr("123456789"),
				Password: stringPtr("wrongpassword"),
			},
			mockUser: func() *domain.User {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				return &domain.User{
					ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Phone:        stringPtr("123456789"),
					PasswordHash: string(hashedPassword),
					IsApproved:   true,
				}
			}(),
			expectError: true,
		},
		{
			name: "error_empty_phone",
			input: &domain.SignInIn{
				Login:    stringPtr(""),
				Password: stringPtr("password"),
			},
			mockUser:     nil,
			repoError:    errors.New("user not found"),
			expectError:  true,
			skipRepoMock: true,
		},
		{
			name: "error_empty_password",
			input: &domain.SignInIn{
				Login:    stringPtr("123456789"),
				Password: stringPtr(""),
			},
			mockUser: func() *domain.User {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				return &domain.User{
					ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Phone:        stringPtr("123456789"),
					PasswordHash: string(hashedPassword),
					IsApproved:   true,
				}
			}(),
			expectError:  true,
			skipRepoMock: true,
		},
		{
			name: "error_session_creation_failed",
			input: &domain.SignInIn{
				Login:    stringPtr("123456789"),
				Password: stringPtr("password"),
			},
			mockUser: func() *domain.User {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				return &domain.User{
					ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Phone:        stringPtr("123456789"),
					PasswordHash: string(hashedPassword),
					IsApproved:   true,
				}
			}(),
			sessionError: errors.New("session creation failed"),
			expectError:  true,
		},
		{
			name: "success_all_metadata",
			input: func() *domain.SignInIn {
				in := &domain.SignInIn{
					Login:    stringPtr("123456789"),
					Password: stringPtr("password"),
				}
				in.Session.DeviceID = uuid.MustParse("00000000-0000-0000-0000-000000000123")
				in.Session.UserAgent = "Mozilla/5.0"
				in.Session.IP = "192.168.1.1"
				return in
			}(),
			mockUser: func() *domain.User {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
				return &domain.User{
					ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Phone:        stringPtr("123456789"),
					PasswordHash: string(hashedPassword),
					IsApproved:   true,
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

			if !tt.skipRepoMock {
				clientRepo.On("GetByPhone", mock.Anything, tt.input.Login).Return(tt.mockUser, tt.repoError)
			}

			// Mock session repository if user exists and password is correct
			if tt.mockUser != nil && tt.repoError == nil && tt.input.Password != nil && *tt.input.Password != "wrongpassword" && *tt.input.Password != "" {
				if tt.sessionError != nil {
					sessionRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Session")).Return(tt.sessionError)
				} else {
					sessionRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Session")).Return(nil)
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
