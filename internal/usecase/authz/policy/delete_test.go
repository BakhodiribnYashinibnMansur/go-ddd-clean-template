package policy_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Delete(t *testing.T) {
	t.Parallel()

	policyID := uuid.New()

	tests := []struct {
		name      string
		id        uuid.UUID
		mockSetup func(*MockPolicyRepo)
		wantErr   bool
	}{
		{
			name: "success",
			id:   policyID,
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Delete", mock.Anything, policyID).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			id:   policyID,
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Delete", mock.Anything, policyID).Return(errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "database error",
			id:   policyID,
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Delete", mock.Anything, policyID).Return(errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name: "nil uuid",
			id:   uuid.Nil,
			mockSetup: func(m *MockPolicyRepo) {
				m.On("Delete", mock.Anything, uuid.Nil).Return(errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, repo := setup(t)
			tt.mockSetup(repo)

			err := uc.Delete(t.Context(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}
