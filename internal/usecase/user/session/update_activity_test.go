package session_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_UpdateActivity_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		filter          *domain.SessionFilter
		mockSession     *domain.Session
		getRepoError    error
		updateRepoError error
		expectError     bool
		validateResult  func(t *testing.T, s *domain.Session)
	}{
		{
			name: "success_update_activity",
			filter: &domain.SessionFilter{
				ID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			mockSession: &domain.Session{
				ID:        func() uuid.UUID { id := uuid.New(); return id }(),
				ExpiresAt: time.Now().Add(time.Hour),
				Revoked:   false,
			},
			getRepoError:    nil,
			updateRepoError: nil,
			expectError:     false,
			validateResult: func(t *testing.T, s *domain.Session) {
				require.False(t, s.LastActivity.IsZero())
				require.False(t, s.UpdatedAt.IsZero())
			},
		},
		{
			name: "error_session_expired",
			filter: &domain.SessionFilter{
				ID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			mockSession: &domain.Session{
				ID:        func() uuid.UUID { id := uuid.New(); return id }(),
				ExpiresAt: time.Now().Add(-time.Hour), // Expired
				Revoked:   false,
			},
			getRepoError:    nil,
			updateRepoError: nil,
			expectError:     true,
		},
		{
			name: "error_session_revoked",
			filter: &domain.SessionFilter{
				ID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			mockSession: &domain.Session{
				ID:        func() uuid.UUID { id := uuid.New(); return id }(),
				ExpiresAt: time.Now().Add(time.Hour),
				Revoked:   true,
			},
			getRepoError:    nil,
			updateRepoError: nil,
			expectError:     true,
		},
		{
			name: "error_session_not_found",
			filter: &domain.SessionFilter{
				ID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			mockSession:     nil,
			getRepoError:    errors.New("session not found"),
			updateRepoError: nil,
			expectError:     true,
		},
		{
			name: "error_get_repository_failure",
			filter: &domain.SessionFilter{
				ID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			mockSession:     nil,
			getRepoError:    errors.New("database error"),
			updateRepoError: nil,
			expectError:     true,
		},
		{
			name: "error_update_repository_failure",
			filter: &domain.SessionFilter{
				ID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			mockSession: &domain.Session{
				ID:        func() uuid.UUID { id := uuid.New(); return id }(),
				ExpiresAt: time.Now().Add(time.Hour),
				Revoked:   false,
			},
			getRepoError:    nil,
			updateRepoError: errors.New("update failed"),
			expectError:     true,
		},
	}

	for _, tt := range tests {
		tt := tt // parallel safety
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			uc, sessionRepo := setup(t)
			ctx := t.Context()

			sessionRepo.On("Get", ctx, tt.filter).Return(tt.mockSession, tt.getRepoError)

			if tt.getRepoError == nil && tt.mockSession != nil && !tt.mockSession.ExpiresAt.Before(time.Now()) && !tt.mockSession.Revoked {
				sessionRepo.On("Update", ctx, mock.MatchedBy(func(s *domain.Session) bool {
					if tt.validateResult != nil {
						tt.validateResult(t, s)
					}
					return true
				})).Return(tt.updateRepoError)
			}

			// act
			err := uc.UpdateActivity(ctx, tt.filter)

			// assert
			if tt.expectError {
				require.Error(t, err)
				if tt.mockSession != nil && tt.mockSession.ExpiresAt.Before(time.Now()) {
					assert.Contains(t, err.Error(), "session invalid or revoked")
				}
			} else {
				require.NoError(t, err)
			}

			sessionRepo.AssertExpectations(t)
		})
	}
}
