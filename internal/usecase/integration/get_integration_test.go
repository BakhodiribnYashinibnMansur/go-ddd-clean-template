package integration_test

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

func TestUseCase_GetIntegration(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	now := time.Now()

	tests := []struct {
		name      string
		id        uuid.UUID
		mockSetup func(*MockRepo)
		wantErr   bool
		check     func(*testing.T, *domain.IntegrationWithKeys)
	}{
		{
			name: "success with api keys",
			id:   id,
			mockSetup: func(m *MockRepo) {
				m.On("GetIntegrationByID", mock.Anything, id).
					Return(&domain.Integration{
						ID:          id,
						Name:        "Stripe",
						Description: "Payment gateway",
						BaseURL:     "https://api.stripe.com",
						IsActive:    true,
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)
				m.On("ListAPIKeysByIntegration", mock.Anything, id).
					Return([]domain.APIKey{
						{ID: uuid.New(), IntegrationID: id, Name: "key-1", IsActive: true},
						{ID: uuid.New(), IntegrationID: id, Name: "key-2", IsActive: false},
					}, nil)
			},
			wantErr: false,
			check: func(t *testing.T, result *domain.IntegrationWithKeys) {
				assert.Equal(t, id, result.ID)
				assert.Equal(t, "Stripe", result.Name)
				assert.Len(t, result.APIKeys, 2)
				assert.Equal(t, "key-1", result.APIKeys[0].Name)
				assert.Equal(t, "key-2", result.APIKeys[1].Name)
			},
		},
		{
			name: "success with no api keys",
			id:   id,
			mockSetup: func(m *MockRepo) {
				m.On("GetIntegrationByID", mock.Anything, id).
					Return(&domain.Integration{
						ID:   id,
						Name: "Empty",
					}, nil)
				m.On("ListAPIKeysByIntegration", mock.Anything, id).
					Return([]domain.APIKey{}, nil)
			},
			wantErr: false,
			check: func(t *testing.T, result *domain.IntegrationWithKeys) {
				assert.Empty(t, result.APIKeys)
			},
		},
		{
			name: "integration not found",
			id:   id,
			mockSetup: func(m *MockRepo) {
				m.On("GetIntegrationByID", mock.Anything, id).
					Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "list api keys error",
			id:   id,
			mockSetup: func(m *MockRepo) {
				m.On("GetIntegrationByID", mock.Anything, id).
					Return(&domain.Integration{ID: id, Name: "Stripe"}, nil)
				m.On("ListAPIKeysByIntegration", mock.Anything, id).
					Return([]domain.APIKey(nil), errors.New("query failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, repo := setup(t)
			tt.mockSetup(repo)

			result, err := uc.GetIntegration(t.Context(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.check != nil {
					tt.check(t, result)
				}
			}

			repo.AssertExpectations(t)
		})
	}
}
