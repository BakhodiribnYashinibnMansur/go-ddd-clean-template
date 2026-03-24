package policy_test

import (
	"errors"
	"testing"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Update(t *testing.T) {
	t.Parallel()

	policyID := uuid.New()
	permID := uuid.New()

	tests := []struct {
		name      string
		policy    *domain.Policy
		mockSetup func(*MockPolicyRepo)
		wantErr   bool
	}{
		{
			name: "success",
			policy: &domain.Policy{
				ID:           policyID,
				PermissionID: permID,
				Effect:       domain.PolicyEffectAllow,
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{"ip": "10.0.0.0/8"},
			},
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.Policy")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success - nil conditions",
			policy: &domain.Policy{
				ID:           policyID,
				PermissionID: permID,
				Effect:       domain.PolicyEffectDeny,
				Priority:     5,
				Active:       false,
				Conditions:   nil,
			},
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.Policy")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "invalid condition key",
			policy: &domain.Policy{
				ID:           policyID,
				PermissionID: permID,
				Effect:       domain.PolicyEffectAllow,
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{"bad_key": "value"},
			},
			mockSetup: func(_ *MockPolicyRepo) {},
			wantErr:   true,
		},
		{
			name: "repo error",
			policy: &domain.Policy{
				ID:           policyID,
				PermissionID: permID,
				Effect:       domain.PolicyEffectAllow,
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{"ip": "10.0.0.0/8"},
			},
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.Policy")).
					Return(errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name: "not found",
			policy: &domain.Policy{
				ID:           policyID,
				PermissionID: permID,
				Effect:       domain.PolicyEffectAllow,
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{"ip": "10.0.0.0/8"},
			},
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.Policy")).
					Return(errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "success - empty conditions map",
			policy: &domain.Policy{
				ID:           policyID,
				PermissionID: permID,
				Effect:       domain.PolicyEffectAllow,
				Priority:     1,
				Active:       true,
				Conditions:   map[string]any{},
			},
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.Policy")).
					Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, repo := setup(t)
			tt.mockSetup(repo)

			err := uc.Update(t.Context(), tt.policy)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}
