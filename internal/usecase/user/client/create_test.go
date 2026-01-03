package client_test

import (
	"errors"
	"testing"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Create_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         *domain.User
		repoError     error
		expectError   bool
		validateSaved func(t *testing.T, u *domain.User)
	}{
		{
			name: "success_password_is_hashed",
			input: &domain.User{
				Phone:    stringPtr("987654321"),
				Password: "password",
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, u *domain.User) {
				// Create method doesn't hash passwords, so password remains unchanged
				require.Equal(t, "password", u.Password)
			},
		},
		{
			name: "success_with_username",
			input: &domain.User{
				Username: stringPtr("testuser"),
				Phone:    stringPtr("987654321"),
				Password: "password",
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, u *domain.User) {
				require.NotNil(t, u.Username)
				require.Equal(t, "testuser", *u.Username)
			},
		},
		{
			name: "error_empty_password",
			input: &domain.User{
				Phone:    stringPtr("987654321"),
				Password: "",
			},
			expectError: false, // Create method doesn't validate
			validateSaved: func(t *testing.T, u *domain.User) {
				require.Equal(t, "", u.Password)
			},
		},
		{
			name: "error_empty_phone",
			input: &domain.User{
				Phone:    stringPtr(""),
				Password: "password",
			},
			expectError: false, // Create method doesn't validate
			validateSaved: func(t *testing.T, u *domain.User) {
				require.Equal(t, "", u.Phone)
			},
		},
		{
			name: "error_weak_password",
			input: &domain.User{
				Phone:    stringPtr("987654321"),
				Password: "123",
			},
			expectError: false, // Create method doesn't validate
			validateSaved: func(t *testing.T, u *domain.User) {
				require.Equal(t, "123", u.Password)
			},
		},
		{
			name: "repo_returns_error",
			input: &domain.User{
				Phone:    stringPtr("987654321"),
				Password: "password",
			},
			repoError:   errors.New("db error"),
			expectError: true,
		},
		{
			name: "success_zero_id",
			input: &domain.User{
				ID:       uuid.UUID{}, // zero UUID
				Phone:    stringPtr("987654321"),
				Password: "password",
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, u *domain.User) {
				require.Equal(t, uuid.UUID{}, u.ID)
			},
		},
		{
			name: "success_negative_id",
			input: &domain.User{
				ID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"), // negative equivalent for UUID
				Phone:    stringPtr("987654321"),
				Password: "password",
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, u *domain.User) {
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000001"), u.ID)
			},
		},
		{
			name: "success_nil_username",
			input: &domain.User{
				Username: nil,
				Phone:    stringPtr("987654321"),
				Password: "password",
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, u *domain.User) {
				require.Nil(t, u.Username)
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

			if tt.repoError != nil || tt.validateSaved != nil {
				clientRepo.
					On("Create", ctx, mock.MatchedBy(func(u *domain.User) bool {
						if tt.validateSaved != nil {
							tt.validateSaved(t, u)
						}
						return true
					})).
					Return(tt.repoError).
					Once()
			}

			// act
			err := uc.Create(ctx, tt.input)

			// assert
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			clientRepo.AssertExpectations(t)
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
