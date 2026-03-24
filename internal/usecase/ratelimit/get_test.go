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

func TestUseCase_GetByID(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	now := time.Now()

	tests := []struct {
		name      string
		id        uuid.UUID
		mockSetup func(*MockRepo)
		wantErr   bool
	}{
		{
			name: "success",
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
			},
			wantErr: false,
		},
		{
			name: "not found error",
			id:   id,
			mockSetup: func(m *MockRepo) {
				m.On("GetByID", mock.Anything, id).
					Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, repo := setup(t)
			tt.mockSetup(repo)

			result, err := uc.GetByID(t.Context(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.id, result.ID)
				assert.Equal(t, "API Rate Limit", result.Name)
			}

			repo.AssertExpectations(t)
		})
	}
}
