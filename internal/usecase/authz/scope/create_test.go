package scope_test

import (
	"errors"
	"testing"

	"gct/internal/domain"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		scope     *domain.Scope
		mockSetup func(*MockScopeRepo)
		wantErr   bool
	}{
		{
			name: "success",
			scope: &domain.Scope{
				Path:   "/api/v1/users",
				Method: "GET",
			},
			mockSetup: func(m *MockScopeRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Scope")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success - POST method",
			scope: &domain.Scope{
				Path:   "/api/v1/users",
				Method: "POST",
			},
			mockSetup: func(m *MockScopeRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Scope")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success - nested path",
			scope: &domain.Scope{
				Path:   "/api/v1/organizations/:orgId/teams/:teamId/members",
				Method: "GET",
			},
			mockSetup: func(m *MockScopeRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Scope")).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repo error",
			scope: &domain.Scope{
				Path:   "/api/v1/users",
				Method: "GET",
			},
			mockSetup: func(m *MockScopeRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Scope")).
					Return(errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name: "duplicate scope error",
			scope: &domain.Scope{
				Path:   "/api/v1/users",
				Method: "GET",
			},
			mockSetup: func(m *MockScopeRepo) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Scope")).
					Return(errors.New("duplicate key value violates unique constraint"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, repo := setup(t)
			tt.mockSetup(repo)

			err := uc.Create(t.Context(), tt.scope)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}
