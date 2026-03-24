package job_test

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
	newName := "Updated Job"
	newCron := "0 0 * * *"

	tests := []struct {
		name      string
		id        uuid.UUID
		req       domain.UpdateJobRequest
		mockSetup func(*MockRepo)
		wantErr   bool
		check     func(*testing.T, *domain.Job)
	}{
		{
			name: "success partial update",
			id:   id,
			req: domain.UpdateJobRequest{
				Name:         &newName,
				CronSchedule: &newCron,
			},
			mockSetup: func(m *MockRepo) {
				m.On("GetByID", mock.Anything, id).
					Return(&domain.Job{
						ID:           id,
						Name:         "Old Job",
						Type:         "cron",
						CronSchedule: "0 * * * *",
						Payload:      map[string]any{"key": "value"},
						IsActive:     true,
						Status:       "idle",
						CreatedAt:    now,
						UpdatedAt:    now,
					}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.Job")).
					Return(nil)
			},
			wantErr: false,
			check: func(t *testing.T, j *domain.Job) {
				assert.Equal(t, newName, j.Name)
				assert.Equal(t, newCron, j.CronSchedule)
				assert.Equal(t, "cron", j.Type)                         // unchanged
				assert.Equal(t, "value", j.Payload["key"])              // unchanged
				assert.True(t, j.IsActive)                              // unchanged
			},
		},
		{
			name: "success update payload",
			id:   id,
			req: domain.UpdateJobRequest{
				Payload: map[string]any{"new_key": "new_value"},
			},
			mockSetup: func(m *MockRepo) {
				m.On("GetByID", mock.Anything, id).
					Return(&domain.Job{
						ID:           id,
						Name:         "Job",
						Type:         "cron",
						CronSchedule: "0 * * * *",
						Payload:      map[string]any{"old_key": "old_value"},
						IsActive:     true,
						Status:       "idle",
						CreatedAt:    now,
						UpdatedAt:    now,
					}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.Job")).
					Return(nil)
			},
			wantErr: false,
			check: func(t *testing.T, j *domain.Job) {
				assert.Equal(t, "new_value", j.Payload["new_key"])
				_, exists := j.Payload["old_key"]
				assert.False(t, exists)
			},
		},
		{
			name: "not found on get",
			id:   id,
			req: domain.UpdateJobRequest{
				Name: &newName,
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
			req: domain.UpdateJobRequest{
				Name: &newName,
			},
			mockSetup: func(m *MockRepo) {
				m.On("GetByID", mock.Anything, id).
					Return(&domain.Job{
						ID:           id,
						Name:         "Old Job",
						Type:         "cron",
						CronSchedule: "0 * * * *",
						Payload:      map[string]any{},
						IsActive:     true,
						Status:       "idle",
						CreatedAt:    now,
						UpdatedAt:    now,
					}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.Job")).
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
