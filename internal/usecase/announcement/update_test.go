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

func TestUseCase_Update(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	now := time.Now()
	newTitle := "Updated Title"
	newContent := "Updated Content"

	tests := []struct {
		name      string
		id        uuid.UUID
		req       domain.UpdateAnnouncementRequest
		mockSetup func(*MockRepo)
		wantErr   bool
		check     func(*testing.T, *domain.Announcement)
	}{
		{
			name: "success partial update",
			id:   id,
			req: domain.UpdateAnnouncementRequest{
				Title:   &newTitle,
				Content: &newContent,
			},
			mockSetup: func(m *MockRepo) {
				m.On("GetByID", mock.Anything, id).
					Return(&domain.Announcement{
						ID:        id,
						Title:     "Old Title",
						Content:   "Old Content",
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
				assert.Equal(t, newTitle, a.Title)
				assert.Equal(t, newContent, a.Content)
				assert.Equal(t, "info", a.Type) // unchanged
				assert.True(t, a.IsActive)       // unchanged
			},
		},
		{
			name: "not found on get",
			id:   id,
			req: domain.UpdateAnnouncementRequest{
				Title: &newTitle,
			},
			mockSetup: func(m *MockRepo) {
				m.On("GetByID", mock.Anything, id).
					Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "repo error on update",
			id:   id,
			req: domain.UpdateAnnouncementRequest{
				Title: &newTitle,
			},
			mockSetup: func(m *MockRepo) {
				m.On("GetByID", mock.Anything, id).
					Return(&domain.Announcement{
						ID:        id,
						Title:     "Old Title",
						Content:   "Old Content",
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

			result, err := uc.Update(t.Context(), tt.id, tt.req)

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
