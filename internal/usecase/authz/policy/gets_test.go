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

func TestUseCase_Gets(t *testing.T) {
	t.Parallel()

	policyID1 := uuid.New()
	policyID2 := uuid.New()
	permID := uuid.New()
	now := time.Now()
	activeTrue := true

	tests := []struct {
		name          string
		filter        *domain.PoliciesFilter
		mockSetup     func(*MockPolicyRepo)
		wantErr       bool
		wantCount     int
		wantPolicies  int
	}{
		{
			name: "success - multiple results",
			filter: &domain.PoliciesFilter{
				PolicyFilter: domain.PolicyFilter{},
			},
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.PoliciesFilter")).
					Return([]*domain.Policy{
						{ID: policyID1, PermissionID: permID, Effect: domain.PolicyEffectAllow, Priority: 10, Active: true, CreatedAt: now},
						{ID: policyID2, PermissionID: permID, Effect: domain.PolicyEffectDeny, Priority: 5, Active: false, CreatedAt: now},
					}, 2, nil)
			},
			wantErr:      false,
			wantCount:    2,
			wantPolicies: 2,
		},
		{
			name: "success - with filter",
			filter: &domain.PoliciesFilter{
				PolicyFilter: domain.PolicyFilter{
					Active: &activeTrue,
				},
			},
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.PoliciesFilter")).
					Return([]*domain.Policy{
						{ID: policyID1, PermissionID: permID, Effect: domain.PolicyEffectAllow, Priority: 10, Active: true, CreatedAt: now},
					}, 1, nil)
			},
			wantErr:      false,
			wantCount:    1,
			wantPolicies: 1,
		},
		{
			name: "success - with pagination",
			filter: &domain.PoliciesFilter{
				PolicyFilter: domain.PolicyFilter{},
				Pagination:   &domain.Pagination{Limit: 10, Offset: 0},
			},
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.PoliciesFilter")).
					Return([]*domain.Policy{
						{ID: policyID1, PermissionID: permID, Effect: domain.PolicyEffectAllow, Priority: 10, Active: true, CreatedAt: now},
					}, 25, nil)
			},
			wantErr:      false,
			wantCount:    25,
			wantPolicies: 1,
		},
		{
			name: "success - empty result",
			filter: &domain.PoliciesFilter{
				PolicyFilter: domain.PolicyFilter{},
			},
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.PoliciesFilter")).
					Return([]*domain.Policy{}, 0, nil)
			},
			wantErr:      false,
			wantCount:    0,
			wantPolicies: 0,
		},
		{
			name: "repo error",
			filter: &domain.PoliciesFilter{
				PolicyFilter: domain.PolicyFilter{},
			},
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.PoliciesFilter")).
					Return([]*domain.Policy(nil), 0, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, repo := setup(t)
			tt.mockSetup(repo)

			policies, count, err := uc.Gets(t.Context(), tt.filter)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, policies)
				assert.Equal(t, 0, count)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCount, count)
				assert.Len(t, policies, tt.wantPolicies)
			}

			repo.AssertExpectations(t)
		})
	}
}
