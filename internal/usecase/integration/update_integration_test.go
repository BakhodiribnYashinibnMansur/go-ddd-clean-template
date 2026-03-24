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

func TestUseCase_UpdateIntegration(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	now := time.Now()
	newName := "Stripe V2"
	newDesc := "Updated description"
	newURL := "https://api.stripe.com/v2"
	newActive := false
	newConfig := map[string]any{"version": "v2"}

	tests := []struct {
		name      string
		id        uuid.UUID
		req       domain.UpdateIntegrationRequest
		mockSetup func(*MockRepo)
		wantErr   bool
		check     func(*testing.T, *domain.Integration)
	}{
		{
			name: "update all fields",
			id:   id,
			req: domain.UpdateIntegrationRequest{
				Name:        &newName,
				Description: &newDesc,
				BaseURL:     &newURL,
				IsActive:    &newActive,
				Config:      &newConfig,
			},
			mockSetup: func(m *MockRepo) {
				m.On("GetIntegrationByID", mock.Anything, id).
					Return(&domain.Integration{
						ID:          id,
						Name:        "Stripe",
						Description: "Payment gateway",
						BaseURL:     "https://api.stripe.com",
						IsActive:    true,
						Config:      map[string]any{"version": "v1"},
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)
				m.On("UpdateIntegration", mock.Anything, mock.AnythingOfType("*domain.Integration")).
					Return(nil)
			},
			wantErr: false,
			check: func(t *testing.T, i *domain.Integration) {
				assert.Equal(t, newName, i.Name)
				assert.Equal(t, newDesc, i.Description)
				assert.Equal(t, newURL, i.BaseURL)
				assert.False(t, i.IsActive)
				assert.Equal(t, newConfig, i.Config)
			},
		},
		{
			name: "partial update - name only",
			id:   id,
			req: domain.UpdateIntegrationRequest{
				Name: &newName,
			},
			mockSetup: func(m *MockRepo) {
				m.On("GetIntegrationByID", mock.Anything, id).
					Return(&domain.Integration{
						ID:       id,
						Name:     "Stripe",
						BaseURL:  "https://api.stripe.com",
						IsActive: true,
					}, nil)
				m.On("UpdateIntegration", mock.Anything, mock.AnythingOfType("*domain.Integration")).
					Return(nil)
			},
			wantErr: false,
			check: func(t *testing.T, i *domain.Integration) {
				assert.Equal(t, newName, i.Name)
				assert.Equal(t, "https://api.stripe.com", i.BaseURL)
				assert.True(t, i.IsActive)
			},
		},
		{
			name: "not found",
			id:   id,
			req:  domain.UpdateIntegrationRequest{Name: &newName},
			mockSetup: func(m *MockRepo) {
				m.On("GetIntegrationByID", mock.Anything, id).
					Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "update repo error",
			id:   id,
			req:  domain.UpdateIntegrationRequest{Name: &newName},
			mockSetup: func(m *MockRepo) {
				m.On("GetIntegrationByID", mock.Anything, id).
					Return(&domain.Integration{ID: id, Name: "Stripe"}, nil)
				m.On("UpdateIntegration", mock.Anything, mock.AnythingOfType("*domain.Integration")).
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

			result, err := uc.UpdateIntegration(t.Context(), tt.id, tt.req)

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
