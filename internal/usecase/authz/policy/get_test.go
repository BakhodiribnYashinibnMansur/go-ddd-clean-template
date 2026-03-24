package policy_test

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

func TestUseCase_Get(t *testing.T) {
	t.Parallel()

	policyID := uuid.New()
	permID := uuid.New()
	now := time.Now()

	tests := []struct {
		name      string
		filter    *domain.PolicyFilter
		mockSetup func(*MockPolicyRepo)
		wantErr   bool
		check     func(*testing.T, *domain.Policy)
	}{
		{
			name: "success - get by id",
			filter: &domain.PolicyFilter{
				ID: &policyID,
			},
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Get", mock.Anything, mock.AnythingOfType("*domain.PolicyFilter")).
					Return(&domain.Policy{
						ID:           policyID,
						PermissionID: permID,
						Effect:       domain.PolicyEffectAllow,
						Priority:     10,
						Active:       true,
						Conditions:   map[string]any{"ip": "10.0.0.0/8"},
						CreatedAt:    now,
					}, nil)
			},
			wantErr: false,
			check: func(t *testing.T, p *domain.Policy) {
				assert.Equal(t, policyID, p.ID)
				assert.Equal(t, permID, p.PermissionID)
				assert.Equal(t, domain.PolicyEffectAllow, p.Effect)
			},
		},
		{
			name: "success - get by permission_id",
			filter: &domain.PolicyFilter{
				PermissionID: &permID,
			},
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Get", mock.Anything, mock.AnythingOfType("*domain.PolicyFilter")).
					Return(&domain.Policy{
						ID:           policyID,
						PermissionID: permID,
						Effect:       domain.PolicyEffectDeny,
						Priority:     5,
						Active:       false,
						CreatedAt:    now,
					}, nil)
			},
			wantErr: false,
			check: func(t *testing.T, p *domain.Policy) {
				assert.Equal(t, permID, p.PermissionID)
				assert.Equal(t, domain.PolicyEffectDeny, p.Effect)
			},
		},
		{
			name: "not found",
			filter: &domain.PolicyFilter{
				ID: &policyID,
			},
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Get", mock.Anything, mock.AnythingOfType("*domain.PolicyFilter")).
					Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "database error",
			filter: &domain.PolicyFilter{
				ID: &policyID,
			},
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Get", mock.Anything, mock.AnythingOfType("*domain.PolicyFilter")).
					Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, repo := setup(t)
			tt.mockSetup(repo)

			result, err := uc.Get(t.Context(), tt.filter)

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
