package session_test

import (
	"errors"
	"testing"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Revoke_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		filter       *domain.SessionFilter
		repoError    error
		expectError  bool
		validateCall func(t *testing.T, filter *domain.SessionFilter)
	}{
		{
			name: "success_revoke_by_id",
			filter: &domain.SessionFilter{
				ID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			repoError:   nil,
			expectError: false,
			validateCall: func(t *testing.T, f *domain.SessionFilter) {
				require.NotNil(t, f.ID)
			},
		},
		{
			name: "success_revoke_by_user_id",
			filter: &domain.SessionFilter{
				UserID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			repoError:   nil,
			expectError: false,
			validateCall: func(t *testing.T, f *domain.SessionFilter) {
				require.NotNil(t, f.UserID)
			},
		},
		{
			name:        "success_revoke_all_sessions",
			filter:      &domain.SessionFilter{}, // Empty filter revokes all
			repoError:   nil,
			expectError: false,
			validateCall: func(t *testing.T, f *domain.SessionFilter) {
				require.Nil(t, f.ID)
				require.Nil(t, f.UserID)
				require.Nil(t, f.Revoked)
			},
		},
		{
			name: "success_revoke_active_sessions",
			filter: &domain.SessionFilter{
				Revoked: func() *bool { r := false; return &r }(),
			},
			repoError:   nil,
			expectError: false,
			validateCall: func(t *testing.T, f *domain.SessionFilter) {
				require.NotNil(t, f.Revoked)
				require.False(t, *f.Revoked)
			},
		},
		{
			name: "success_revoke_already_revoked",
			filter: &domain.SessionFilter{
				Revoked: func() *bool { r := true; return &r }(),
			},
			repoError:   nil,
			expectError: false,
			validateCall: func(t *testing.T, f *domain.SessionFilter) {
				require.NotNil(t, f.Revoked)
				require.True(t, *f.Revoked)
			},
		},
		{
			name: "success_revoke_multiple_filters",
			filter: &domain.SessionFilter{
				ID:     func() *uuid.UUID { id := uuid.New(); return &id }(),
				UserID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			repoError:   nil,
			expectError: false,
			validateCall: func(t *testing.T, f *domain.SessionFilter) {
				require.NotNil(t, f.ID)
				require.NotNil(t, f.UserID)
			},
		},
		{
			name: "error_repository_failure",
			filter: &domain.SessionFilter{
				ID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			repoError:   errors.New("database error"),
			expectError: true,
		},
		{
			name:        "error_nil_filter",
			filter:      nil,
			repoError:   errors.New("invalid filter"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt // parallel safety
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			uc, sessionRepo := setup(t)
			ctx := t.Context()

			sessionRepo.On("Revoke", ctx, mock.MatchedBy(func(f *domain.SessionFilter) bool {
				if tt.validateCall != nil {
					tt.validateCall(t, f)
				}
				return true
			})).Return(tt.repoError).Once()

			// act
			err := uc.Revoke(ctx, tt.filter)

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
