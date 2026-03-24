package usersetting_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Set(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name      string
		userID    uuid.UUID
		key       string
		value     string
		mockSetup func(*MockRepo)
		wantErr   bool
	}{
		{
			name:   "success",
			userID: userID,
			key:    "theme",
			value:  "dark",
			mockSetup: func(m *MockRepo) {
				m.On("Upsert", mock.Anything, mock.AnythingOfType("*domain.UserSetting")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "repo error",
			userID: userID,
			key:    "theme",
			value:  "dark",
			mockSetup: func(m *MockRepo) {
				m.On("Upsert", mock.Anything, mock.AnythingOfType("*domain.UserSetting")).Return(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:   "empty key",
			userID: userID,
			key:    "",
			value:  "val",
			mockSetup: func(m *MockRepo) {
				m.On("Upsert", mock.Anything, mock.AnythingOfType("*domain.UserSetting")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "empty value",
			userID: userID,
			key:    "theme",
			value:  "",
			mockSetup: func(m *MockRepo) {
				m.On("Upsert", mock.Anything, mock.AnythingOfType("*domain.UserSetting")).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, repo := setup(t)
			tt.mockSetup(repo)

			err := uc.Set(t.Context(), tt.userID, tt.key, tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			repo.AssertExpectations(t)
		})
	}
}
