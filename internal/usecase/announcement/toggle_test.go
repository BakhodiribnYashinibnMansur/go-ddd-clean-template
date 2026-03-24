package announcement_test

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

func TestUseCase_Toggle(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	now := time.Now()

	tests := []struct {
		name      string
		id        uuid.UUID
		mockSetup func(*MockRepo)
		wantErr   bool
		check     func(*testing.T, *domain.Announcement)
	}{
		{
			name: "active to inactive",
			id:   id,
			mockSetup: func(m *MockRepo) {
				m.On("GetByID", mock.Anything, id).
					Return(&domain.Announcement{
						ID:        id,
						Title:     "Test",
						Content:   "Content",
						Type:      "info",
						IsActive:  true,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.Announcement")).
					Return(nil)
			},
			wantErr: false,
			check: func(t *testing.T, a *domain.Announcement) {
				assert.False(t, a.IsActive)
			},
		},
		{
			name: "inactive to active",
			id:   id,
			mockSetup: func(m *MockRepo) {
				m.On("GetByID", mock.Anything, id).
					Return(&domain.Announcement{
						ID:        id,
						Title:     "Test",
						Content:   "Content",
						Type:      "info",
						IsActive:  false,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.Announcement")).
					Return(nil)
			},
			wantErr: false,
			check: func(t *testing.T, a *domain.Announcement) {
				assert.True(t, a.IsActive)
			},
		},
		{
			name: "not found",
			id:   id,
			mockSetup: func(m *MockRepo) {
				m.On("GetByID", mock.Anything, id).
					Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "update error",
			id:   id,
			mockSetup: func(m *MockRepo) {
				m.On("GetByID", mock.Anything, id).
					Return(&domain.Announcement{
						ID:        id,
						Title:     "Test",
						Content:   "Content",
						Type:      "info",
						IsActive:  true,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.Announcement")).
					Return(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, repo := setup(t)
			tt.mockSetup(repo)

			result, err := uc.Toggle(t.Context(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				if tt.check != nil {
					tt.check(t, result)
				}
			}

			repo.AssertExpectations(t)
		})
	}
}
