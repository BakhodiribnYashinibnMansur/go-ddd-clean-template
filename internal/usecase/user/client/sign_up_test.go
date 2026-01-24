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

func TestUseCase_SignUp_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         *domain.SignUpIn
		repoError     error
		expectError   bool
		validateSaved func(t *testing.T, u *domain.User)
	}{
		{
			name: "success_basic_signup",
			input: &domain.SignUpIn{
				Username: "testuser",
				Phone:    "123456789",
				Password: "Password123!",
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, u *domain.User) {
				t.Helper()
				require.NotNil(t, u.Username)
				require.Equal(t, "testuser", *u.Username)
				require.Equal(t, "123456789", *u.Phone)
				// Password should be hashed by SignUp method
				require.NotEmpty(t, u.PasswordHash)
				require.NotEqual(t, "Password123!", u.PasswordHash)
			},
		},
		{
			name: "success_empty_username",
			input: &domain.SignUpIn{
				Username: "",
				Phone:    "123456789",
				Password: "Password123!",
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, u *domain.User) {
				t.Helper()
				require.Nil(t, u.Username)
				require.Equal(t, "123456789", *u.Phone)
			},
		},
		{
			name: "error_empty_phone",
			input: &domain.SignUpIn{
				Username: "testuser",
				Phone:    "",
				Password: "Password123!",
			},
			repoError:   nil,
			expectError: true,
		},
		{
			name: "error_empty_password",
			input: &domain.SignUpIn{
				Username: "testuser",
				Phone:    "123456789",
				Password: "",
			},
			repoError:   nil,
			expectError: true,
		},
		{
			name: "error_weak_password",
			input: &domain.SignUpIn{
				Username: "testuser",
				Phone:    "123456789",
				Password: "123",
			},
			repoError:   nil,
			expectError: true,
		},
		{
			name: "error_repository_failure",
			input: &domain.SignUpIn{
				Username: "testuser",
				Phone:    "123456789",
				Password: "Password123!",
			},
			repoError:   errors.New("database error"),
			expectError: true,
		},
		{
			name: "success_long_username",
			input: &domain.SignUpIn{
				Username: "verylongusernamethatmightstillwork",
				Phone:    "123456789",
				Password: "Password123!",
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, u *domain.User) {
				t.Helper()
				require.NotNil(t, u.Username)
				require.Equal(t, "verylongusernamethatmightstillwork", *u.Username)
			},
		},
		{
			name: "success_special_chars_password",
			input: &domain.SignUpIn{
				Username: "testuser",
				Phone:    "123456789",
				Password: "P@ssw0rd!1",
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, u *domain.User) {
				t.Helper()
				require.Equal(t, "123456789", *u.Phone)
				require.NotNil(t, u.Username)
				require.Equal(t, "testuser", *u.Username)
				// Special characters should be handled and hashed
				require.NotEmpty(t, u.PasswordHash)
			},
		},
		{
			name: "success_numeric_phone",
			input: &domain.SignUpIn{
				Username: "testuser",
				Phone:    "9876543210",
				Password: "Password123!",
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, u *domain.User) {
				t.Helper()
				require.Equal(t, "9876543210", *u.Phone)
				require.NotNil(t, u.Username)
				require.Equal(t, "testuser", *u.Username)
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

			if tt.repoError != nil || tt.validateSaved != nil {
				clientRepo.
					On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
						if tt.validateSaved != nil {
							tt.validateSaved(t, u)
						}
						return true
					})).
					Return(tt.repoError).
					Once()

				if tt.repoError == nil {
					// SignUp calls SignIn which calls GetByPhone and Session Create
					hash, _ := bcrypt.GenerateFromPassword([]byte(tt.input.Password), bcrypt.DefaultCost)
					clientRepo.On("GetByPhone", mock.Anything, tt.input.Phone).
						Return(&domain.User{
							ID:           uuid.New(),
							Phone:        &tt.input.Phone,
							PasswordHash: string(hash),
							IsApproved:   true,
						}, nil).Once()

					sessionRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Session")).
						Return(nil).Once()
				}
			}

			// act
			res, err := uc.SignUp(ctx, tt.input)

			// assert
			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, res)
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.NotEmpty(t, res.AccessToken)
			}

			clientRepo.AssertExpectations(t)
		})
	}
}
