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

func TestUseCase_ListIntegrations(t *testing.T) {
	t.Parallel()

	now := time.Now()
	active := true

	tests := []struct {
		name      string
		filter    domain.IntegrationFilter
		mockSetup func(*MockRepo)
		wantErr   bool
		wantLen   int
		wantTotal int64
	}{
		{
			name:   "success with results",
			filter: domain.IntegrationFilter{Limit: 10, Offset: 0},
			mockSetup: func(m *MockRepo) {
				m.On("ListIntegrations", mock.Anything, domain.IntegrationFilter{Limit: 10, Offset: 0}).
					Return([]domain.Integration{
						{ID: uuid.New(), Name: "Stripe", IsActive: true, CreatedAt: now, UpdatedAt: now},
						{ID: uuid.New(), Name: "PayPal", IsActive: false, CreatedAt: now, UpdatedAt: now},
					}, int64(2), nil)
			},
			wantErr:   false,
			wantLen:   2,
			wantTotal: 2,
		},
		{
			name:   "success empty results",
			filter: domain.IntegrationFilter{Limit: 10},
			mockSetup: func(m *MockRepo) {
				m.On("ListIntegrations", mock.Anything, domain.IntegrationFilter{Limit: 10}).
					Return([]domain.Integration{}, int64(0), nil)
			},
			wantErr:   false,
			wantLen:   0,
			wantTotal: 0,
		},
		{
			name:   "filter by active",
			filter: domain.IntegrationFilter{IsActive: &active, Limit: 10},
			mockSetup: func(m *MockRepo) {
				m.On("ListIntegrations", mock.Anything, domain.IntegrationFilter{IsActive: &active, Limit: 10}).
					Return([]domain.Integration{
						{ID: uuid.New(), Name: "Stripe", IsActive: true},
					}, int64(1), nil)
			},
			wantErr:   false,
			wantLen:   1,
			wantTotal: 1,
		},
		{
			name:   "filter by search",
			filter: domain.IntegrationFilter{Search: "stripe", Limit: 10},
			mockSetup: func(m *MockRepo) {
				m.On("ListIntegrations", mock.Anything, domain.IntegrationFilter{Search: "stripe", Limit: 10}).
					Return([]domain.Integration{
						{ID: uuid.New(), Name: "Stripe"},
					}, int64(1), nil)
			},
			wantErr:   false,
			wantLen:   1,
			wantTotal: 1,
		},
		{
			name:   "repo error",
			filter: domain.IntegrationFilter{Limit: 10},
			mockSetup: func(m *MockRepo) {
				m.On("ListIntegrations", mock.Anything, domain.IntegrationFilter{Limit: 10}).
					Return([]domain.Integration(nil), int64(0), errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, repo := setup(t)
			tt.mockSetup(repo)

			items, total, err := uc.ListIntegrations(t.Context(), tt.filter)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, items, tt.wantLen)
				assert.Equal(t, tt.wantTotal, total)
			}

			repo.AssertExpectations(t)
		})
	}
}
