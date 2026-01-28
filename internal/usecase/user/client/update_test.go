package client_test

import (
	"errors"
	"testing"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Update_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         *domain.User
		repoError     error
		expectError   bool
		validateSaved func(t *testing.T, u *domain.User)
		setupExisting func(u *domain.User)
	}{
		{
			name: "success_basic_update",
			input: &domain.User{
				ID:    uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Phone: stringPtr("111222333"),
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, u *domain.User) {
				t.Helper()
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000001"), u.ID)
				require.Equal(t, "111222333", *u.Phone)
			},
		},
		{
			name: "success_with_username",
			input: &domain.User{
				ID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Username: stringPtr("updateduser"),
				Phone:    stringPtr("111222333"),
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, u *domain.User) {
				t.Helper()
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000001"), u.ID)
				require.NotNil(t, u.Username)
				require.Equal(t, "updateduser", *u.Username)
			},
		},
		{
			name: "success_with_password",
			input: &domain.User{
				ID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Phone:    stringPtr("111222333"),
				Password: "Password123!",
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, u *domain.User) {
				t.Helper()
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000001"), u.ID)
				require.Equal(t, "Password123!", u.Password)
			},
		},
		{
			name: "success_empty_phone",
			input: &domain.User{
				ID:    uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Phone: stringPtr(""),
			},
			repoError:   nil,
			expectError: false,
			setupExisting: func(u *domain.User) {
				u.Phone = stringPtr("111222333")
			},
			validateSaved: func(t *testing.T, u *domain.User) {
				t.Helper()
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000001"), u.ID)
				require.Equal(t, "111222333", *u.Phone)
			},
		},
		{
			name: "success_nil_username",
			input: &domain.User{
				ID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Username: nil,
				Phone:    stringPtr("111222333"),
			},
			repoError:   nil,
			expectError: false,
			setupExisting: func(u *domain.User) {
				u.Username = nil
			},
			validateSaved: func(t *testing.T, u *domain.User) {
				t.Helper()
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000001"), u.ID)
				require.Nil(t, u.Username)
			},
		},
		{
			name: "success_zero_id",
			input: &domain.User{
				ID:    uuid.UUID{}, // zero UUID
				Phone: stringPtr("111222333"),
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, u *domain.User) {
				t.Helper()
				require.Equal(t, uuid.UUID{}, u.ID)
			},
		},
		{
			name: "success_negative_id",
			input: &domain.User{
				ID:    uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"), // max UUID equivalent
				Phone: stringPtr("111222333"),
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, u *domain.User) {
				t.Helper()
				require.Equal(t, uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"), u.ID)
			},
		},
		{
			name: "error_repository_failure",
			input: &domain.User{
				ID:    uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Phone: stringPtr("111222333"),
			},
			repoError:   errors.New("database error"),
			expectError: true,
		},
		{
			name: "success_all_fields",
			input: &domain.User{
				ID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"),
				Username: stringPtr("fulluser"),
				Phone:    stringPtr("999888777"),
				Password: "Password123!",
			},
			repoError:   nil,
			expectError: false,
			setupExisting: func(u *domain.User) {
				u.Username = stringPtr("olduser")
				u.Phone = stringPtr("111111111")
				u.Password = "oldpass"
			},
			validateSaved: func(t *testing.T, u *domain.User) {
				t.Helper()
				require.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000001"), u.ID)
				require.NotNil(t, u.Username)
				require.Equal(t, "fulluser", *u.Username)
				require.Equal(t, "999888777", *u.Phone)
				require.Equal(t, "Password123!", u.Password)
			},
		},
	}

	for _, tt := range tests {
		// parallel safety
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			uc, clientRepo, _ := setup(t)
			ctx := t.Context()

			// Mock Get (Update now calls Get to fetch existing user)
			existingUser := &domain.User{
				ID:       tt.input.ID,
				Phone:    stringPtr("old-phone"),
				Username: stringPtr("old-username"),
			}
			if tt.setupExisting != nil {
				tt.setupExisting(existingUser)
			}
			clientRepo.On("Get", mock.Anything, mock.MatchedBy(func(f *domain.UserFilter) bool {
				return f.ID != nil && *f.ID == tt.input.ID
			})).Return(existingUser, nil).Maybe()

			if tt.repoError != nil || tt.validateSaved != nil {
				clientRepo.
					On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
						if tt.validateSaved != nil {
							tt.validateSaved(t, u)
						}
						return true
					})).
					Return(tt.repoError).
					Once()
			}

			// act
			err := uc.Update(ctx, tt.input)

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
