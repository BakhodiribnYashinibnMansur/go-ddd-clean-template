package usersetting_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Gets(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name      string
		userID    uuid.UUID
		mockSetup func(*MockRepo)
		want      []domain.UserSetting
		wantErr   bool
	}{
		{
			name:   "success - returns settings",
			userID: userID,
			mockSetup: func(m *MockRepo) {
				m.On("Gets", t.Context(), userID).Return([]domain.UserSetting{
					{ID: uuid.New(), UserID: userID, Key: "theme", Value: "dark", CreatedAt: now, UpdatedAt: now},
					{ID: uuid.New(), UserID: userID, Key: "lang", Value: "en", CreatedAt: now, UpdatedAt: now},
				}, nil)
			},
			want: []domain.UserSetting{
				{UserID: userID, Key: "theme", Value: "dark"},
				{UserID: userID, Key: "lang", Value: "en"},
			},
			wantErr: false,
		},
		{
			name:   "success - empty list",
			userID: userID,
			mockSetup: func(m *MockRepo) {
				m.On("Gets", t.Context(), userID).Return([]domain.UserSetting(nil), nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "repo error",
			userID: userID,
			mockSetup: func(m *MockRepo) {
				m.On("Gets", t.Context(), userID).Return(nil, errors.New("db error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, repo := setup(t)
			tt.mockSetup(repo)

			got, err := uc.Gets(t.Context(), tt.userID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, len(tt.want))
			repo.AssertExpectations(t)
		})
	}
}
