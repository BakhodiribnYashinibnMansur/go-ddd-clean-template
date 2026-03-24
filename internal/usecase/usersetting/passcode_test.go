package usersetting_test

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

func TestUseCase_SetPasscode(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name      string
		userID    uuid.UUID
		passcode  string
		mockSetup func(*MockRepo)
		wantErr   bool
	}{
		{
			name:     "success",
			userID:   userID,
			passcode: "1234",
			mockSetup: func(m *MockRepo) {
				m.On("Upsert", mock.Anything, mock.AnythingOfType("*domain.UserSetting")).Return(nil).Twice()
			},
			wantErr: false,
		},
		{
			name:     "first upsert fails",
			userID:   userID,
			passcode: "1234",
			mockSetup: func(m *MockRepo) {
				m.On("Upsert", mock.Anything, mock.AnythingOfType("*domain.UserSetting")).Return(errors.New("db error")).Once()
			},
			wantErr: true,
		},
		{
			name:     "second upsert fails",
			userID:   userID,
			passcode: "1234",
			mockSetup: func(m *MockRepo) {
				m.On("Upsert", mock.Anything, mock.AnythingOfType("*domain.UserSetting")).Return(nil).Once()
				m.On("Upsert", mock.Anything, mock.AnythingOfType("*domain.UserSetting")).Return(errors.New("db error")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, repo := setup(t)
			tt.mockSetup(repo)

			err := uc.SetPasscode(t.Context(), tt.userID, tt.passcode)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestUseCase_VerifyPasscode(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name      string
		userID    uuid.UUID
		passcode  string
		mockSetup func(*MockRepo)
		want      bool
		wantErr   bool
	}{
		{
			name:     "correct passcode",
			userID:   userID,
			passcode: "1234",
			mockSetup: func(m *MockRepo) {
				m.On("Gets", mock.Anything, userID).Return([]domain.UserSetting{
					{ID: uuid.New(), UserID: userID, Key: "passcode", Value: "1234", CreatedAt: now, UpdatedAt: now},
					{ID: uuid.New(), UserID: userID, Key: "passcode_enabled", Value: "true", CreatedAt: now, UpdatedAt: now},
				}, nil)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:     "wrong passcode",
			userID:   userID,
			passcode: "9999",
			mockSetup: func(m *MockRepo) {
				m.On("Gets", mock.Anything, userID).Return([]domain.UserSetting{
					{ID: uuid.New(), UserID: userID, Key: "passcode", Value: "1234", CreatedAt: now, UpdatedAt: now},
				}, nil)
			},
			want:    false,
			wantErr: false,
		},
		{
			name:     "no passcode set",
			userID:   userID,
			passcode: "1234",
			mockSetup: func(m *MockRepo) {
				m.On("Gets", mock.Anything, userID).Return([]domain.UserSetting{}, nil)
			},
			want:    false,
			wantErr: false,
		},
		{
			name:     "repo error",
			userID:   userID,
			passcode: "1234",
			mockSetup: func(m *MockRepo) {
				m.On("Gets", mock.Anything, userID).Return(nil, errors.New("db error"))
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, repo := setup(t)
			tt.mockSetup(repo)

			got, err := uc.VerifyPasscode(t.Context(), tt.userID, tt.passcode)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			repo.AssertExpectations(t)
		})
	}
}

func TestUseCase_RemovePasscode(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name      string
		userID    uuid.UUID
		mockSetup func(*MockRepo)
		wantErr   bool
	}{
		{
			name:   "success",
			userID: userID,
			mockSetup: func(m *MockRepo) {
				m.On("Delete", mock.Anything, userID, "passcode").Return(nil)
				m.On("Delete", mock.Anything, userID, "passcode_enabled").Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "delete errors are ignored",
			userID: userID,
			mockSetup: func(m *MockRepo) {
				m.On("Delete", mock.Anything, userID, "passcode").Return(errors.New("db error"))
				m.On("Delete", mock.Anything, userID, "passcode_enabled").Return(errors.New("db error"))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, repo := setup(t)
			tt.mockSetup(repo)

			err := uc.RemovePasscode(t.Context(), tt.userID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			repo.AssertExpectations(t)
		})
	}
}
