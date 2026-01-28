package session_test

import (
	"errors"
	"testing"

	"gct/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Delete_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		filter       *domain.SessionFilter
		repoError    error
		expectError  bool
		validateCall func(t *testing.T, filter *domain.SessionFilter)
	}{
		{
			name: "success_delete_by_id",
			filter: &domain.SessionFilter{
				ID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			repoError:   nil,
			expectError: false,
			validateCall: func(t *testing.T, f *domain.SessionFilter) {
				t.Helper()
				require.NotNil(t, f.ID)
			},
		},
		{
			name: "success_delete_by_user_id",
			filter: &domain.SessionFilter{
				UserID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			repoError:   nil,
			expectError: false,
		},
		{
			name: "success_delete_with_revoked_true",
			filter: &domain.SessionFilter{
				Revoked: func() *bool { r := true; return &r }(),
			},
			repoError:   nil,
			expectError: false,
			validateCall: func(t *testing.T, f *domain.SessionFilter) {
				t.Helper()
				require.NotNil(t, f.Revoked)
				require.True(t, *f.Revoked)
			},
		},
		{
			name: "success_delete_with_revoked_false",
			filter: &domain.SessionFilter{
				Revoked: func() *bool { r := false; return &r }(),
			},
			repoError:   nil,
			expectError: false,
			validateCall: func(t *testing.T, f *domain.SessionFilter) {
				t.Helper()
				require.NotNil(t, f.Revoked)
				require.False(t, *f.Revoked)
			},
		},
		{
			name: "success_delete_with_multiple_filters",
			filter: &domain.SessionFilter{
				ID:     func() *uuid.UUID { id := uuid.New(); return &id }(),
				UserID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			repoError:   nil,
			expectError: false,
			validateCall: func(t *testing.T, f *domain.SessionFilter) {
				t.Helper()
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
		{
			name:        "success_empty_filter",
			filter:      &domain.SessionFilter{},
			repoError:   nil,
			expectError: false,
			validateCall: func(t *testing.T, f *domain.SessionFilter) {
				t.Helper()
				require.Nil(t, f.ID)
				require.Nil(t, f.UserID)
				require.Nil(t, f.Revoked)
			},
		},
	}

	for _, tt := range tests {
		// parallel safety
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			uc, sessionRepo := setup(t)
			ctx := t.Context()

			sessionRepo.On("Delete", ctx, mock.MatchedBy(func(f *domain.SessionFilter) bool {
				if tt.validateCall != nil {
					tt.validateCall(t, f)
				}
				return true
			})).Return(tt.repoError).Once()

			// act
			err := uc.Delete(ctx, tt.filter)

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
