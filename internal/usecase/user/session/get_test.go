package session_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Get_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		filter         *domain.SessionFilter
		mockSession    *domain.Session
		repoError      error
		expectError    bool
		expectDelete   bool
		validateResult func(t *testing.T, s *domain.Session)
	}{
		{
			name: "success_get_by_id",
			filter: &domain.SessionFilter{
				ID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			mockSession: &domain.Session{
				ID:        func() uuid.UUID { id := uuid.New(); return id }(),
				UserID:    uuid.New(),
				ExpiresAt: time.Now().Add(time.Hour),
			},
			repoError:    nil,
			expectError:  false,
			expectDelete: false,
			validateResult: func(t *testing.T, s *domain.Session) {
				t.Helper()
				require.NotEqual(t, uuid.Nil, s.ID)
				require.NotEqual(t, uuid.Nil, s.UserID)
				require.True(t, s.ExpiresAt.After(time.Now()))
			},
		},
		{
			name: "success_get_by_user_id",
			filter: &domain.SessionFilter{
				UserID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			mockSession: &domain.Session{
				ID:        func() uuid.UUID { id := uuid.New(); return id }(),
				UserID:    uuid.New(),
				ExpiresAt: time.Now().Add(time.Hour),
			},
			repoError:    nil,
			expectError:  false,
			expectDelete: false,
		},
		{
			name: "success_get_active_session",
			filter: &domain.SessionFilter{
				Revoked: func() *bool { r := false; return &r }(),
			},
			mockSession: &domain.Session{
				ID:        func() uuid.UUID { id := uuid.New(); return id }(),
				Revoked:   false,
				ExpiresAt: time.Now().Add(time.Hour),
			},
			repoError:    nil,
			expectError:  false,
			expectDelete: false,
			validateResult: func(t *testing.T, s *domain.Session) {
				t.Helper()
				require.False(t, s.Revoked)
			},
		},
		{
			name: "success_get_revoked_session",
			filter: &domain.SessionFilter{
				Revoked: func() *bool { r := true; return &r }(),
			},
			mockSession: &domain.Session{
				ID:        func() uuid.UUID { id := uuid.New(); return id }(),
				Revoked:   true,
				ExpiresAt: time.Now().Add(time.Hour),
			},
			repoError:    nil,
			expectError:  false,
			expectDelete: false,
			validateResult: func(t *testing.T, s *domain.Session) {
				t.Helper()
				require.True(t, s.Revoked)
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
			},
			repoError:    nil,
			expectError:  true,
			expectDelete: true,
		},
		{
			name: "error_session_not_found",
			filter: &domain.SessionFilter{
				ID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			mockSession:  nil,
			repoError:    errors.New("session not found"),
			expectError:  true,
			expectDelete: false,
		},
		{
			name: "error_repository_failure",
			filter: &domain.SessionFilter{
				ID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			mockSession:  nil,
			repoError:    errors.New("database error"),
			expectError:  true,
			expectDelete: false,
		},
		{
			name: "success_get_with_ip_address",
			filter: &domain.SessionFilter{
				ID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			mockSession: &domain.Session{
				ID:        func() uuid.UUID { id := uuid.New(); return id }(),
				IPAddress: func() *string { ip := "192.168.1.1"; return &ip }(),
				ExpiresAt: time.Now().Add(time.Hour),
			},
			repoError:    nil,
			expectError:  false,
			expectDelete: false,
			validateResult: func(t *testing.T, s *domain.Session) {
				t.Helper()
				require.NotNil(t, s.IPAddress)
				require.Equal(t, "192.168.1.1", *s.IPAddress)
			},
		},
		{
			name: "success_get_with_user_agent",
			filter: &domain.SessionFilter{
				ID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			mockSession: &domain.Session{
				ID:        func() uuid.UUID { id := uuid.New(); return id }(),
				UserAgent: func() *string { ua := "Mozilla/5.0"; return &ua }(),
				ExpiresAt: time.Now().Add(time.Hour),
			},
			repoError:    nil,
			expectError:  false,
			expectDelete: false,
			validateResult: func(t *testing.T, s *domain.Session) {
				t.Helper()
				require.NotNil(t, s.UserAgent)
				require.Equal(t, "Mozilla/5.0", *s.UserAgent)
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

			sessionRepo.On("Get", ctx, tt.filter).Return(tt.mockSession, tt.repoError)

			// If session is expired, expect Delete to be called
			if tt.expectDelete {
				sessionRepo.On("Delete", ctx, tt.filter).Return(nil)
			}

			// act
			out, err := uc.Get(ctx, tt.filter)

			// assert
			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, out)
				if tt.mockSession != nil && tt.mockSession.ExpiresAt.Before(time.Now()) {
					assert.Contains(t, err.Error(), "session expired")
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, out)
				if tt.validateResult != nil {
					tt.validateResult(t, out)
				}
			}

			sessionRepo.AssertExpectations(t)
		})
	}
}
