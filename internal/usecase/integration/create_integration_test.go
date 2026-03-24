package integration_test

import (
	"errors"
	"testing"

	"gct/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_CreateIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		req       domain.CreateIntegrationRequest
		mockSetup func(*MockRepo)
		wantErr   bool
		check     func(*testing.T, *domain.Integration)
	}{
		{
			name: "success",
			req: domain.CreateIntegrationRequest{
				Name:        "Stripe",
				Description: "Payment gateway",
				BaseURL:     "https://api.stripe.com",
				IsActive:    true,
				Config:      map[string]any{"version": "v1"},
			},
			mockSetup: func(m *MockRepo) {
				m.On("CreateIntegration", mock.Anything, mock.AnythingOfType("*domain.Integration")).
					Return(nil)
			},
			wantErr: false,
			check: func(t *testing.T, i *domain.Integration) {
				assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", i.ID.String())
				assert.Equal(t, "Stripe", i.Name)
				assert.Equal(t, "Payment gateway", i.Description)
				assert.Equal(t, "https://api.stripe.com", i.BaseURL)
				assert.True(t, i.IsActive)
				assert.Equal(t, map[string]any{"version": "v1"}, i.Config)
			},
		},
		{
			name: "success with nil config",
			req: domain.CreateIntegrationRequest{
				Name:     "PayPal",
				BaseURL:  "https://api.paypal.com",
				IsActive: false,
			},
			mockSetup: func(m *MockRepo) {
				m.On("CreateIntegration", mock.Anything, mock.AnythingOfType("*domain.Integration")).
					Return(nil)
			},
			wantErr: false,
			check: func(t *testing.T, i *domain.Integration) {
				assert.Equal(t, "PayPal", i.Name)
				assert.False(t, i.IsActive)
			},
		},
		{
			name: "repo error",
			req: domain.CreateIntegrationRequest{
				Name:    "FailIntegration",
				BaseURL: "https://fail.example.com",
			},
			mockSetup: func(m *MockRepo) {
				m.On("CreateIntegration", mock.Anything, mock.AnythingOfType("*domain.Integration")).
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

			result, err := uc.CreateIntegration(t.Context(), tt.req)

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
