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

func TestUseCase_Toggle(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	now := time.Now()

	tests := []struct {
		name      string
		id        uuid.UUID
		mockSetup func(*MockRepo)
		wantErr   bool
		check     func(*testing.T, *domain.RateLimit)
	}{
		{
			name: "active to inactive",
			id:   id,
			mockSetup: func(m *MockRepo) {
				m.On("GetByID", mock.Anything, id).
					Return(&domain.RateLimit{
						ID:            id,
						Name:          "API Rate Limit",
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
				assert.False(t, rl.IsActive)
			},
		},
		{
			name: "inactive to active",
			id:   id,
			mockSetup: func(m *MockRepo) {
				m.On("GetByID", mock.Anything, id).
					Return(&domain.RateLimit{
						ID:            id,
						Name:          "API Rate Limit",
						PathPattern:   "/api/v1/*",
						Method:        "GET",
						LimitCount:    100,
						WindowSeconds: 60,
						IsActive:      false,
						CreatedAt:     now,
						UpdatedAt:     now,
					}, nil)
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.RateLimit")).
					Return(nil)
			},
			wantErr: false,
			check: func(t *testing.T, rl *domain.RateLimit) {
				assert.True(t, rl.IsActive)
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
					Return(&domain.RateLimit{
						ID:            id,
						Name:          "API Rate Limit",
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
