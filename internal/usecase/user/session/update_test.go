package session_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Update_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         *domain.Session
		repoError     error
		expectError   bool
		validateSaved func(t *testing.T, s *domain.Session)
	}{
		{
			name: "success_basic_update",
			input: &domain.Session{
				ID:     func() uuid.UUID { id := uuid.New(); return id }(),
				UserID: uuid.New(),
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, s *domain.Session) {
				t.Helper()
				require.NotEqual(t, uuid.Nil, s.ID)
				require.NotEqual(t, uuid.Nil, s.UserID)
			},
		},
		{
			name: "success_update_with_device_name",
			input: &domain.Session{
				ID:         func() uuid.UUID { id := uuid.New(); return id }(),
				UserID:     uuid.New(),
				DeviceName: func() *string { name := "iPhone 13"; return &name }(),
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, s *domain.Session) {
				t.Helper()
				require.NotNil(t, s.DeviceName)
				require.Equal(t, "iPhone 13", *s.DeviceName)
			},
		},
		{
			name: "success_update_with_device_type",
			input: &domain.Session{
				ID:         func() uuid.UUID { id := uuid.New(); return id }(),
				UserID:     uuid.New(),
				DeviceType: func() *domain.SessionDeviceType { dt := domain.DeviceTypeMobile; return &dt }(),
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, s *domain.Session) {
				t.Helper()
				require.NotNil(t, s.DeviceType)
				require.Equal(t, domain.DeviceTypeMobile, *s.DeviceType)
			},
		},
		{
			name: "success_update_with_ip_address",
			input: &domain.Session{
				ID:        func() uuid.UUID { id := uuid.New(); return id }(),
				UserID:    uuid.New(),
				IPAddress: func() *string { ip := "192.168.1.100"; return &ip }(),
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, s *domain.Session) {
				t.Helper()
				require.NotNil(t, s.IPAddress)
				require.Equal(t, "192.168.1.100", *s.IPAddress)
			},
		},
		{
			name: "success_update_with_user_agent",
			input: &domain.Session{
				ID:        func() uuid.UUID { id := uuid.New(); return id }(),
				UserID:    uuid.New(),
				UserAgent: func() *string { ua := "Mozilla/5.0 (iPhone)"; return &ua }(),
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, s *domain.Session) {
				t.Helper()
				require.NotNil(t, s.UserAgent)
				require.Equal(t, "Mozilla/5.0 (iPhone)", *s.UserAgent)
			},
		},
		{
			name: "success_update_with_fcm_token",
			input: &domain.Session{
				ID:       func() uuid.UUID { id := uuid.New(); return id }(),
				UserID:   uuid.New(),
				FCMToken: func() *string { token := "fcm_token_123"; return &token }(),
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, s *domain.Session) {
				t.Helper()
				require.NotNil(t, s.FCMToken)
				require.Equal(t, "fcm_token_123", *s.FCMToken)
			},
		},
		{
			name: "success_update_with_expires_at",
			input: &domain.Session{
				ID:        func() uuid.UUID { id := uuid.New(); return id }(),
				UserID:    uuid.New(),
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, s *domain.Session) {
				t.Helper()
				require.True(t, s.ExpiresAt.After(time.Now()))
			},
		},
		{
			name: "success_update_revoked_status",
			input: &domain.Session{
				ID:      func() uuid.UUID { id := uuid.New(); return id }(),
				UserID:  uuid.New(),
				Revoked: true,
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, s *domain.Session) {
				t.Helper()
				require.True(t, s.Revoked)
			},
		},
		{
			name: "success_update_all_fields",
			input: &domain.Session{
				ID:               func() uuid.UUID { id := uuid.New(); return id }(),
				DeviceID:         func() uuid.UUID { id := uuid.New(); return id }(),
				DeviceName:       func() *string { name := "Samsung Galaxy"; return &name }(),
				DeviceType:       func() *domain.SessionDeviceType { dt := domain.DeviceTypeMobile; return &dt }(),
				IPAddress:        func() *string { ip := "192.168.1.50"; return &ip }(),
				UserAgent:        func() *string { ua := "Chrome/96.0"; return &ua }(),
				FCMToken:         func() *string { token := "fcm_token_456"; return &token }(),
				UserID:           uuid.New(),
				RefreshTokenHash: "refresh_hash_123",
				ExpiresAt:        time.Now().Add(48 * time.Hour),
				Revoked:          false,
			},
			repoError:   nil,
			expectError: false,
			validateSaved: func(t *testing.T, s *domain.Session) {
				t.Helper()
				require.NotEqual(t, uuid.Nil, s.ID)
				require.NotEqual(t, uuid.Nil, s.DeviceID)
				require.NotNil(t, s.DeviceName)
				require.NotNil(t, s.DeviceType)
				require.NotNil(t, s.IPAddress)
				require.NotNil(t, s.UserAgent)
				require.NotNil(t, s.FCMToken)
				require.NotEqual(t, uuid.Nil, s.UserID)
				require.Equal(t, "refresh_hash_123", s.RefreshTokenHash)
				require.False(t, s.Revoked)
			},
		},
		{
			name: "error_repository_failure",
			input: &domain.Session{
				ID:     func() uuid.UUID { id := uuid.New(); return id }(),
				UserID: uuid.New(),
			},
			repoError:   errors.New("database error"),
			expectError: true,
		},
		{
			name:        "error_nil_input",
			input:       nil,
			repoError:   errors.New("invalid session"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		// parallel safety
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// arrange
			uc, sessionRepo := setup(t)
			ctx := t.Context()

			if tt.repoError != nil || tt.validateSaved != nil {
				sessionRepo.
					On("Update", ctx, mock.MatchedBy(func(s *domain.Session) bool {
						if tt.validateSaved != nil {
							tt.validateSaved(t, s)
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

			sessionRepo.AssertExpectations(t)
		})
	}
}
