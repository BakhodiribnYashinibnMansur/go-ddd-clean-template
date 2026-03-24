package usersetting_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Delete(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name      string
		userID    uuid.UUID
		key       string
		mockSetup func(*MockRepo)
		wantErr   bool
	}{
		{
			name:   "success",
			userID: userID,
			key:    "theme",
			mockSetup: func(m *MockRepo) {
				m.On("Delete", t.Context(), userID, "theme").Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "repo error",
			userID: userID,
			key:    "theme",
			mockSetup: func(m *MockRepo) {
				m.On("Delete", t.Context(), userID, "theme").Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, repo := setup(t)
			tt.mockSetup(repo)

			err := uc.Delete(t.Context(), tt.userID, tt.key)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			repo.AssertExpectations(t)
		})
	}
}
