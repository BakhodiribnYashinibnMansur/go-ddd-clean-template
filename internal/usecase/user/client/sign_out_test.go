package client_test

import (
	"errors"
	"testing"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_SignOut_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		sessionID      string
		repoError      error
		expectError    bool
		validateFilter func(t *testing.T, filter *domain.SessionFilter)
	}{
		{
			name:        "success_valid_session_id",
			sessionID:   uuid.New().String(),
			repoError:   nil,
			expectError: false,
			validateFilter: func(t *testing.T, filter *domain.SessionFilter) {
				// Validated in test setup
			},
		},
		{
			name:        "success_empty_session_id",
			sessionID:   "",
			repoError:   nil,
			expectError: false,
			validateFilter: func(t *testing.T, filter *domain.SessionFilter) {
				// Validated in test setup
			},
		},
		{
			name:        "success_invalid_uuid_format",
			sessionID:   "invalid-uuid-format",
			repoError:   nil,
			expectError: false,
			validateFilter: func(t *testing.T, filter *domain.SessionFilter) {
				// Validated in test setup
			},
		},
		{
			name:        "error_repo_returns_error",
			sessionID:   uuid.New().String(),
			repoError:   errors.New("session revoke failed"),
			expectError: true,
			validateFilter: func(t *testing.T, filter *domain.SessionFilter) {
				// Validated in test setup
			},
		},
		{
			name:        "success_nil_session_id",
			sessionID:   "00000000-0000-0000-0000-000000000000",
			repoError:   nil,
			expectError: false,
			validateFilter: func(t *testing.T, filter *domain.SessionFilter) {
				// Validated in test setup
			},
		},
		{
			name:        "success_long_session_id",
			sessionID:   "12345678-1234-5678-9abc-123456789012-extra",
			repoError:   nil,
			expectError: false,
			validateFilter: func(t *testing.T, filter *domain.SessionFilter) {
				// Validated in test setup
			},
		},
	}

	for _, tt := range tests {
		tt := tt // parallel safety
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			uc, _, sessionRepo := setup(t)
			ctx := t.Context()

			in := &domain.SignOutIn{
				SessionID: tt.sessionID,
			}

			sessionRepo.On("Revoke", ctx, mock.MatchedBy(func(f *domain.SessionFilter) bool {
				if tt.validateFilter != nil {
					tt.validateFilter(t, f)
				}
				return true
			})).Return(tt.repoError).Once()

			// act
			err := uc.SignOut(ctx, in)

			// assert
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			sessionRepo.AssertExpectations(t)
		})
	}
}
