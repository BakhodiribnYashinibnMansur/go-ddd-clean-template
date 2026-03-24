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

func TestUseCase_List(t *testing.T) {
	t.Parallel()

	now := time.Now()
	active := true

	tests := []struct {
		name      string
		filter    domain.RateLimitFilter
		mockSetup func(*MockRepo)
		wantCount int64
		wantLen   int
		wantErr   bool
	}{
		{
			name: "success with items",
			filter: domain.RateLimitFilter{
				Limit:    10,
				Offset:   0,
				IsActive: &active,
			},
			mockSetup: func(m *MockRepo) {
				m.On("List", mock.Anything, domain.RateLimitFilter{
					Limit:    10,
					Offset:   0,
					IsActive: &active,
				}).Return([]domain.RateLimit{
					{
						ID:            uuid.New(),
						Name:          "Rate Limit 1",
						PathPattern:   "/api/v1/*",
						Method:        "GET",
						LimitCount:    100,
						WindowSeconds: 60,
						IsActive:      true,
						CreatedAt:     now,
						UpdatedAt:     now,
					},
					{
						ID:            uuid.New(),
						Name:          "Rate Limit 2",
						PathPattern:   "/api/v2/*",
						Method:        "POST",
						LimitCount:    50,
						WindowSeconds: 30,
						IsActive:      true,
						CreatedAt:     now,
						UpdatedAt:     now,
					},
				}, int64(2), nil)
			},
			wantCount: 2,
			wantLen:   2,
			wantErr:   false,
		},
		{
			name:   "empty list",
			filter: domain.RateLimitFilter{Limit: 10},
			mockSetup: func(m *MockRepo) {
				m.On("List", mock.Anything, domain.RateLimitFilter{Limit: 10}).
					Return([]domain.RateLimit{}, int64(0), nil)
			},
			wantCount: 0,
			wantLen:   0,
			wantErr:   false,
		},
		{
			name:   "repo error",
			filter: domain.RateLimitFilter{Limit: 10},
			mockSetup: func(m *MockRepo) {
				m.On("List", mock.Anything, domain.RateLimitFilter{Limit: 10}).
					Return([]domain.RateLimit{}, int64(0), errors.New("database error"))
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
