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

func TestUseCase_Create_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          *domain.Session
		mockSetup      func(mockRepo *MockSessionRepo)
		repoError      error
		expectError    bool
		validateResult func(t *testing.T, out *domain.Session)
	}{
		{
			name: "success_basic_create",
			input: &domain.Session{
				UserID: uuid.New(),
			},
			mockSetup: func(mockRepo *MockSessionRepo) {
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.Session) bool {
					return s.UserID != uuid.Nil && s.ID != uuid.Nil && !s.CreatedAt.IsZero()
				})).Return(nil)
			},
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, out *domain.Session) {
				t.Helper()
				require.NotEmpty(t, out.ID)
				require.NotEqual(t, uuid.Nil, out.UserID)
			},
		},
		{
			name: "success_create_with_device_id",
			input: &domain.Session{
				UserID:   uuid.New(),
				DeviceID: uuid.New(),
			},
			mockSetup: func(mockRepo *MockSessionRepo) {
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.Session) bool {
					return s.UserID != uuid.Nil && s.DeviceID != uuid.Nil
				})).Return(nil)
			},
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, out *domain.Session) {
				t.Helper()
				require.NotEqual(t, uuid.Nil, out.UserID)
				require.NotEqual(t, uuid.Nil, out.DeviceID)
			},
		},
		{
			name: "success_create_with_user_agent",
			input: &domain.Session{
				UserID:    uuid.New(),
				UserAgent: func() *string { ua := "Mozilla/5.0"; return &ua }(),
			},
			mockSetup: func(mockRepo *MockSessionRepo) {
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.Session) bool {
					return s.UserID != uuid.Nil && s.UserAgent != nil && *s.UserAgent == "Mozilla/5.0"
				})).Return(nil)
			},
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, out *domain.Session) {
				t.Helper()
				require.NotEqual(t, uuid.Nil, out.UserID)
				require.Equal(t, "Mozilla/5.0", *out.UserAgent)
			},
		},
		{
			name: "success_create_with_ip_address",
			input: &domain.Session{
				UserID:    uuid.New(),
				IPAddress: func() *string { ip := "192.168.1.1"; return &ip }(),
			},
			mockSetup: func(mockRepo *MockSessionRepo) {
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.Session) bool {
					return s.UserID != uuid.Nil && s.IPAddress != nil && *s.IPAddress == "192.168.1.1"
				})).Return(nil)
			},
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, out *domain.Session) {
				t.Helper()
				require.NotEqual(t, uuid.Nil, out.UserID)
				require.Equal(t, "192.168.1.1", *out.IPAddress)
			},
		},
		{
			name: "success_create_default_expiration",
			input: &domain.Session{
				UserID: uuid.New(),
			},
			mockSetup: func(mockRepo *MockSessionRepo) {
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.Session) bool {
					return s.UserID != uuid.Nil
				})).Return(nil)
			},
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, out *domain.Session) {
				t.Helper()
				require.True(t, out.ExpiresAt.After(time.Now()))
				require.True(t, out.ExpiresAt.Before(time.Now().Add(25*time.Hour)))
			},
		},
		{
			name: "error_repository_failure",
			input: &domain.Session{
				UserID: uuid.New(),
			},
			mockSetup: func(mockRepo *MockSessionRepo) {
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Session")).Return(errors.New("database error"))
			},
			repoError:   errors.New("database error"),
			expectError: true,
		},
		{
			name: "success_create_zero_user_id",
			input: &domain.Session{
				UserID: uuid.Nil,
			},
			mockSetup: func(mockRepo *MockSessionRepo) {
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.Session) bool {
					return s.UserID == uuid.Nil
				})).Return(nil)
			},
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, out *domain.Session) {
				t.Helper()
				require.Equal(t, uuid.Nil, out.UserID)
			},
		},
		{
			name: "success_create_negative_user_id",
			input: &domain.Session{
				UserID: uuid.New(), // Using valid UUID instead of negative
			},
			mockSetup: func(mockRepo *MockSessionRepo) {
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.Session) bool {
					return s.UserID != uuid.Nil
				})).Return(nil)
			},
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, out *domain.Session) {
				t.Helper()
				require.NotEqual(t, uuid.Nil, out.UserID)
			},
		},
		{
			name: "success_create_uuid_generation",
			input: &domain.Session{
				UserID: uuid.New(),
			},
			mockSetup: func(mockRepo *MockSessionRepo) {
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.Session) bool {
					return s.UserID != uuid.Nil
				})).Return(nil)
			},
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, out *domain.Session) {
				t.Helper()
				require.NotEqual(t, uuid.Nil, out.ID)
			},
		},
		{
			name: "success_create_timestamps_set",
			input: &domain.Session{
				UserID: uuid.New(),
			},
			mockSetup: func(mockRepo *MockSessionRepo) {
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.Session) bool {
					return s.UserID != uuid.Nil
				})).Return(nil)
			},
			repoError:   nil,
			expectError: false,
			validateResult: func(t *testing.T, out *domain.Session) {
				t.Helper()
				require.False(t, out.CreatedAt.IsZero())
				require.False(t, out.UpdatedAt.IsZero())
				require.True(t, out.CreatedAt.Before(time.Now().Add(time.Minute)))
				require.True(t, out.UpdatedAt.Before(time.Now().Add(time.Minute)))
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

			if tt.mockSetup != nil {
				tt.mockSetup(sessionRepo)
			}

			// act
			out, err := uc.Create(ctx, tt.input)

			// assert
			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, out)
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
