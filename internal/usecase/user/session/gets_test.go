package session_test

import (
	"errors"
	"testing"

	"gct/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Gets_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		filter         *domain.SessionsFilter
		mockSessions   []*domain.Session
		mockTotal      int
		repoError      error
		expectError    bool
		validateResult func(t *testing.T, sessions []*domain.Session, total int)
	}{
		{
			name: "success_basic_gets",
			filter: &domain.SessionsFilter{
				Pagination: &domain.Pagination{Limit: 10, Offset: 0},
			},
			mockSessions: []*domain.Session{
				{UserID: uuid.New(), ID: func() uuid.UUID { id := uuid.New(); return id }()},
				{UserID: uuid.New(), ID: func() uuid.UUID { id := uuid.New(); return id }()},
			},
			mockTotal:   2,
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, sessions []*domain.Session, total int) {
				t.Helper()
				require.Len(t, sessions, 2)
				require.Equal(t, 2, total)
				require.NotEqual(t, uuid.Nil, sessions[0].UserID)
				require.NotEqual(t, uuid.Nil, sessions[1].UserID)
			},
		},
		{
			name: "success_gets_with_pagination",
			filter: &domain.SessionsFilter{
				Pagination: &domain.Pagination{Limit: 5, Offset: 10},
			},
			mockSessions: []*domain.Session{
				{UserID: uuid.New(), ID: func() uuid.UUID { id := uuid.New(); return id }()},
				{UserID: uuid.New(), ID: func() uuid.UUID { id := uuid.New(); return id }()},
				{UserID: uuid.New(), ID: func() uuid.UUID { id := uuid.New(); return id }()},
				{UserID: uuid.New(), ID: func() uuid.UUID { id := uuid.New(); return id }()},
				{UserID: uuid.New(), ID: func() uuid.UUID { id := uuid.New(); return id }()},
			},
			mockTotal:   25,
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, sessions []*domain.Session, total int) {
				t.Helper()
				require.Len(t, sessions, 5)
				require.Equal(t, 25, total)
			},
		},
		{
			name: "success_gets_by_user_id",
			filter: &domain.SessionsFilter{
				SessionFilter: domain.SessionFilter{
					UserID: func() *uuid.UUID { id := uuid.New(); return &id }(),
				},
				Pagination: &domain.Pagination{Limit: 10, Offset: 0},
			},
			mockSessions: []*domain.Session{
				{UserID: uuid.New(), ID: func() uuid.UUID { id := uuid.New(); return id }()},
				{UserID: uuid.New(), ID: func() uuid.UUID { id := uuid.New(); return id }()},
			},
			mockTotal:   2,
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, sessions []*domain.Session, total int) {
				t.Helper()
				require.Len(t, sessions, 2)
				for _, session := range sessions {
					require.NotEqual(t, uuid.Nil, session.UserID)
				}
			},
		},
		{
			name: "success_gets_active_sessions",
			filter: &domain.SessionsFilter{
				SessionFilter: domain.SessionFilter{
					Revoked: func() *bool { r := false; return &r }(),
				},
				Pagination: &domain.Pagination{Limit: 10, Offset: 0},
			},
			mockSessions: []*domain.Session{
				{UserID: uuid.New(), Revoked: false, ID: func() uuid.UUID { id := uuid.New(); return id }()},
				{UserID: uuid.New(), Revoked: false, ID: func() uuid.UUID { id := uuid.New(); return id }()},
			},
			mockTotal:   2,
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, sessions []*domain.Session, total int) {
				t.Helper()
				require.Len(t, sessions, 2)
				for _, session := range sessions {
					require.False(t, session.Revoked)
				}
			},
		},
		{
			name: "success_gets_revoked_sessions",
			filter: &domain.SessionsFilter{
				SessionFilter: domain.SessionFilter{
					Revoked: func() *bool { r := true; return &r }(),
				},
				Pagination: &domain.Pagination{Limit: 10, Offset: 0},
			},
			mockSessions: []*domain.Session{
				{UserID: uuid.New(), Revoked: true, ID: func() uuid.UUID { id := uuid.New(); return id }()},
				{UserID: uuid.New(), Revoked: true, ID: func() uuid.UUID { id := uuid.New(); return id }()},
			},
			mockTotal:   2,
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, sessions []*domain.Session, total int) {
				t.Helper()
				require.Len(t, sessions, 2)
				for _, session := range sessions {
					require.True(t, session.Revoked)
				}
			},
		},
		{
			name: "success_gets_empty_result",
			filter: &domain.SessionsFilter{
				Pagination: &domain.Pagination{Limit: 10, Offset: 0},
			},
			mockSessions: []*domain.Session{},
			mockTotal:    0,
			repoError:    nil,
			expectError:  false,
			validateResult: func(t *testing.T, sessions []*domain.Session, total int) {
				t.Helper()
				require.Empty(t, sessions)
				require.Equal(t, 0, total)
			},
		},
		{
			name: "error_repository_failure",
			filter: &domain.SessionsFilter{
				Pagination: &domain.Pagination{Limit: 10, Offset: 0},
			},
			mockSessions: nil,
			mockTotal:    0,
			repoError:    errors.New("database error"),
			expectError:  true,
		},
		{
			name: "success_gets_with_ip_addresses",
			filter: &domain.SessionsFilter{
				Pagination: &domain.Pagination{Limit: 5, Offset: 0},
			},
			mockSessions: []*domain.Session{
				{UserID: uuid.New(), IPAddress: func() *string { ip := "192.168.1.1"; return &ip }(), ID: func() uuid.UUID { id := uuid.New(); return id }()},
				{UserID: uuid.New(), IPAddress: func() *string { ip := "192.168.1.2"; return &ip }(), ID: func() uuid.UUID { id := uuid.New(); return id }()},
			},
			mockTotal:   2,
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, sessions []*domain.Session, total int) {
				t.Helper()
				require.Len(t, sessions, 2)
				for _, session := range sessions {
					require.NotNil(t, session.IPAddress)
				}
			},
		},
		{
			name: "success_gets_multiple_filters",
			filter: &domain.SessionsFilter{
				SessionFilter: domain.SessionFilter{
					UserID:  func() *uuid.UUID { id := uuid.New(); return &id }(),
					Revoked: func() *bool { r := false; return &r }(),
				},
				Pagination: &domain.Pagination{Limit: 10, Offset: 0},
			},
			mockSessions: []*domain.Session{
				{UserID: uuid.New(), Revoked: false, ID: func() uuid.UUID { id := uuid.New(); return id }()},
			},
			mockTotal:   1,
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, sessions []*domain.Session, total int) {
				t.Helper()
				require.Len(t, sessions, 1)
				require.NotEqual(t, uuid.Nil, sessions[0].UserID)
				require.False(t, sessions[0].Revoked)
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

			sessionRepo.On("Gets", ctx, tt.filter).Return(tt.mockSessions, tt.mockTotal, tt.repoError)

			// act
			out, total, err := uc.Gets(ctx, tt.filter)

			// assert
			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, out)
				assert.Equal(t, 0, total)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, out)
				if tt.validateResult != nil {
					tt.validateResult(t, out, total)
				}
			}

			sessionRepo.AssertExpectations(t)
		})
	}
}
