package announcement_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Create(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name      string
		req       domain.CreateAnnouncementRequest
		mockSetup func(*MockRepo)
		wantErr   bool
	}{
		{
			name: "success",
			req: domain.CreateAnnouncementRequest{
				Title:    "Maintenance Notice",
				Content:  "Scheduled maintenance this weekend.",
				Type:     "info",
				IsActive: true,
				StartsAt: &now,
				EndsAt:   nil,
			},
			mockSetup: func(m *MockRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Announcement")).
					Run(func(args mock.Arguments) {
						a := args.Get(1).(*domain.Announcement)
						assert.NotEmpty(t, a.ID)
						assert.Equal(t, "Maintenance Notice", a.Title)
						assert.Equal(t, "Scheduled maintenance this weekend.", a.Content)
						assert.Equal(t, "info", a.Type)
						assert.True(t, a.IsActive)
						assert.NotNil(t, a.StartsAt)
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repo error",
			req: domain.CreateAnnouncementRequest{
				Title:    "Maintenance Notice",
				Content:  "Content",
				Type:     "warning",
				IsActive: false,
			},
			mockSetup: func(m *MockRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Announcement")).
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

			result, err := uc.Create(t.Context(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.Title, result.Title)
				assert.Equal(t, tt.req.Content, result.Content)
				assert.Equal(t, tt.req.Type, result.Type)
				assert.Equal(t, tt.req.IsActive, result.IsActive)
			}

			repo.AssertExpectations(t)
		})
	}
}
