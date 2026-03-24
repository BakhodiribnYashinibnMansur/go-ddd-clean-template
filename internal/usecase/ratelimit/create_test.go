package ratelimit_test

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
		req       domain.CreateRateLimitRequest
		mockSetup func(*MockRepo)
		wantErr   bool
	}{
		{
			name: "success",
			req: domain.CreateRateLimitRequest{
				Name:          "API Rate Limit",
				PathPattern:   "/api/v1/*",
				Method:        "GET",
				LimitCount:    100,
				WindowSeconds: 60,
				IsActive:      true,
			},
			mockSetup: func(m *MockRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.RateLimit")).
					Run(func(args mock.Arguments) {
						rl := args.Get(1).(*domain.RateLimit)
						assert.NotEmpty(t, rl.ID)
						assert.Equal(t, "API Rate Limit", rl.Name)
						assert.Equal(t, "/api/v1/*", rl.PathPattern)
						assert.Equal(t, "GET", rl.Method)
						assert.Equal(t, 100, rl.LimitCount)
						assert.Equal(t, 60, rl.WindowSeconds)
						assert.True(t, rl.IsActive)
					}).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repo error",
			req: domain.CreateRateLimitRequest{
				Name:          "API Rate Limit",
				PathPattern:   "/api/v1/*",
				Method:        "POST",
				LimitCount:    50,
				WindowSeconds: 30,
				IsActive:      false,
			},
			mockSetup: func(m *MockRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.RateLimit")).
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
				assert.Equal(t, tt.req.PathPattern, result.PathPattern)
				assert.Equal(t, tt.req.Method, result.Method)
				assert.Equal(t, tt.req.LimitCount, result.LimitCount)
				assert.Equal(t, tt.req.WindowSeconds, result.WindowSeconds)
				assert.Equal(t, tt.req.IsActive, result.IsActive)
			}

			repo.AssertExpectations(t)
		})
	}
}
