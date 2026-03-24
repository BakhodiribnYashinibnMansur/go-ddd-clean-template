package policy_test

import (
	"errors"
	"testing"

	"gct/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Create(t *testing.T) {
	t.Parallel()

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
				PermissionID: permID,
				Effect:       domain.PolicyEffectAllow,
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{"ip": "192.168.1.0/24"},
			},
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Policy")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success - nil conditions",
			policy: &domain.Policy{
				PermissionID: permID,
				Effect:       domain.PolicyEffectDeny,
				Priority:     5,
				Active:       false,
				Conditions:   nil,
			},
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Policy")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "invalid condition key",
			policy: &domain.Policy{
				PermissionID: permID,
				Effect:       domain.PolicyEffectAllow,
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{"invalid_key_xyz": "value"},
			},
			mockSetup: func(_ *MockPolicyRepo) {},
			wantErr:   true,
		},
		{
			name: "repo error",
			policy: &domain.Policy{
				PermissionID: permID,
				Effect:       domain.PolicyEffectAllow,
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{"ip": "10.0.0.0/8"},
			},
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Policy")).
					Return(errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name: "success - multiple valid condition keys",
			policy: &domain.Policy{
				PermissionID: permID,
				Effect:       domain.PolicyEffectAllow,
				Priority:     100,
				Active:       true,
				Conditions: map[string]any{
					"ip":         "10.0.0.0/8",
					"user_agent": "Mozilla",
					"time":       "09:00-17:00",
				},
			},
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Policy")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "empty conditions map",
			policy: &domain.Policy{
				PermissionID: permID,
				Effect:       domain.PolicyEffectAllow,
				Priority:     1,
				Active:       true,
				Conditions:   map[string]any{},
			},
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Policy")).
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

			err := uc.Create(t.Context(), tt.policy)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}
