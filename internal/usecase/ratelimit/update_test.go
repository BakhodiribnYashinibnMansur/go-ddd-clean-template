package ratelimit_test

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
	newName := "Updated Rate Limit"
	newPath := "/api/v2/*"
	newLimit := 200

	tests := []struct {
		name      string
		id        uuid.UUID
		req       domain.UpdateRateLimitRequest
		mockSetup func(*MockRepo)
		wantErr   bool
		check     func(*testing.T, *domain.RateLimit)
	}{
		{
			name: "success partial update",
			id:   id,
			req: domain.UpdateRateLimitRequest{
				Name:        &newName,
				PathPattern: &newPath,
				LimitCount:  &newLimit,
			},
			mockSetup: func(m *MockRepo) {
				m.On("GetByID", mock.Anything, id).
					Return(&domain.RateLimit{
						ID:            id,
						Name:          "Old Rate Limit",
						PathPattern:   "/api/v1/*",
						Method:        "GET",
						LimitCount:    100,
						WindowSeconds: 60,
						IsActive:      true,
						CreatedAt:     now,
						UpdatedAt:     now,
					}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.RateLimit")).
					Return(nil)
			},
			wantErr: false,
			check: func(t *testing.T, rl *domain.RateLimit) {
				assert.Equal(t, newName, rl.Name)
				assert.Equal(t, newPath, rl.PathPattern)
				assert.Equal(t, newLimit, rl.LimitCount)
				assert.Equal(t, "GET", rl.Method)     // unchanged
				assert.Equal(t, 60, rl.WindowSeconds)  // unchanged
				assert.True(t, rl.IsActive)             // unchanged
			},
		},
		{
			name: "not found on get",
			id:   id,
			req: domain.UpdateRateLimitRequest{
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
			req: domain.UpdateRateLimitRequest{
				Name: &newName,
			},
			mockSetup: func(m *MockRepo) {
				m.On("GetByID", mock.Anything, id).
					Return(&domain.RateLimit{
						ID:            id,
						Name:          "Old Rate Limit",
						PathPattern:   "/api/v1/*",
						Method:        "GET",
						LimitCount:    100,
						WindowSeconds: 60,
						IsActive:      true,
						CreatedAt:     now,
						UpdatedAt:     now,
					}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.RateLimit")).
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
