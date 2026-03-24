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

func TestUseCase_List(t *testing.T) {
	t.Parallel()

	now := time.Now()
	active := true

	tests := []struct {
		name       string
		filter     domain.AnnouncementFilter
		mockSetup  func(*MockRepo)
		wantCount  int64
		wantLen    int
		wantErr    bool
	}{
		{
			name: "success with items",
			filter: domain.AnnouncementFilter{
				Limit:    10,
				Offset:   0,
				IsActive: &active,
			},
			mockSetup: func(m *MockRepo) {
				m.On("List", mock.Anything, domain.AnnouncementFilter{
					Limit:    10,
					Offset:   0,
					IsActive: &active,
				}).Return([]domain.Announcement{
					{
						ID:        uuid.New(),
						Title:     "First",
						Content:   "Content 1",
						Type:      "info",
						IsActive:  true,
						CreatedAt: now,
						UpdatedAt: now,
					},
					{
						ID:        uuid.New(),
						Title:     "Second",
						Content:   "Content 2",
						Type:      "warning",
						IsActive:  true,
						CreatedAt: now,
						UpdatedAt: now,
					},
				}, int64(2), nil)
			},
			wantCount: 2,
			wantLen:   2,
			wantErr:   false,
		},
		{
			name:   "empty list",
			filter: domain.AnnouncementFilter{Limit: 10},
			mockSetup: func(m *MockRepo) {
				m.On("List", mock.Anything, domain.AnnouncementFilter{Limit: 10}).
					Return([]domain.Announcement{}, int64(0), nil)
			},
			wantCount: 0,
			wantLen:   0,
			wantErr:   false,
		},
		{
			name:   "repo error",
			filter: domain.AnnouncementFilter{Limit: 10},
			mockSetup: func(m *MockRepo) {
				m.On("List", mock.Anything, domain.AnnouncementFilter{Limit: 10}).
					Return([]domain.Announcement{}, int64(0), errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, repo := setup(t)
			tt.mockSetup(repo)

			items, total, err := uc.List(t.Context(), tt.filter)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCount, total)
				assert.Len(t, items, tt.wantLen)
			}

			repo.AssertExpectations(t)
		})
	}
}
