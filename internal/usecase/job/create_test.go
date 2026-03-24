package job_test

import (
	"errors"
	"testing"

	"gct/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		req       domain.CreateJobRequest
		mockSetup func(*MockRepo)
		wantErr   bool
	}{
		{
			name: "success",
			req: domain.CreateJobRequest{
				Name:         "Email Sender",
				Type:         "cron",
				CronSchedule: "0 * * * *",
				Payload:      map[string]any{"template": "welcome"},
				IsActive:     true,
			},
			mockSetup: func(m *MockRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Job")).
					Run(func(args mock.Arguments) {
						j := args.Get(1).(*domain.Job)
						assert.NotEmpty(t, j.ID)
						assert.Equal(t, "Email Sender", j.Name)
						assert.Equal(t, "cron", j.Type)
						assert.Equal(t, "0 * * * *", j.CronSchedule)
						assert.Equal(t, "idle", j.Status)
						assert.True(t, j.IsActive)
						assert.NotNil(t, j.Payload)
						assert.Equal(t, "welcome", j.Payload["template"])
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success with nil payload defaults to empty map",
			req: domain.CreateJobRequest{
				Name:         "Cleanup Job",
				Type:         "cron",
				CronSchedule: "0 0 * * *",
				Payload:      nil,
				IsActive:     true,
			},
			mockSetup: func(m *MockRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Job")).
					Run(func(args mock.Arguments) {
						j := args.Get(1).(*domain.Job)
						assert.NotNil(t, j.Payload)
						assert.Empty(t, j.Payload)
						assert.Equal(t, "idle", j.Status)
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repo error",
			req: domain.CreateJobRequest{
				Name:         "Email Sender",
				Type:         "cron",
				CronSchedule: "0 * * * *",
				Payload:      map[string]any{},
				IsActive:     false,
			},
			mockSetup: func(m *MockRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Job")).
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
				assert.Equal(t, tt.req.Name, result.Name)
				assert.Equal(t, tt.req.Type, result.Type)
				assert.Equal(t, tt.req.CronSchedule, result.CronSchedule)
				assert.Equal(t, "idle", result.Status)
				assert.NotNil(t, result.Payload)
			}

			repo.AssertExpectations(t)
		})
	}
}
