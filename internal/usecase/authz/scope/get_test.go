package scope_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Get(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name      string
		filter    *domain.ScopeFilter
		mockSetup func(*MockScopeRepo)
		wantErr   bool
		check     func(*testing.T, *domain.Scope)
	}{
		{
			name: "success - get by path",
			filter: &domain.ScopeFilter{
				Path: func() *string { s := "/api/v1/users"; return &s }(),
			},
			mockSetup: func(m *MockScopeRepo) {
				m.On("Get", mock.Anything, mock.AnythingOfType("*domain.ScopeFilter")).
					Return(&domain.Scope{
						Path:      "/api/v1/users",
						Method:    "GET",
						CreatedAt: now,
					}, nil)
			},
			wantErr: false,
			check: func(t *testing.T, s *domain.Scope) {
				assert.Equal(t, "/api/v1/users", s.Path)
				assert.Equal(t, "GET", s.Method)
			},
		},
		{
			name: "success - get by path and method",
			filter: &domain.ScopeFilter{
				Path:   func() *string { s := "/api/v1/users"; return &s }(),
				Method: func() *string { s := "POST"; return &s }(),
			},
			mockSetup: func(m *MockScopeRepo) {
				m.On("Get", mock.Anything, mock.AnythingOfType("*domain.ScopeFilter")).
					Return(&domain.Scope{
						Path:      "/api/v1/users",
						Method:    "POST",
						CreatedAt: now,
					}, nil)
			},
			wantErr: false,
			check: func(t *testing.T, s *domain.Scope) {
				assert.Equal(t, "/api/v1/users", s.Path)
				assert.Equal(t, "POST", s.Method)
			},
		},
		{
			name: "not found",
			filter: &domain.ScopeFilter{
				Path: func() *string { s := "/api/v1/nonexistent"; return &s }(),
			},
			mockSetup: func(m *MockScopeRepo) {
				m.On("Get", mock.Anything, mock.AnythingOfType("*domain.ScopeFilter")).
					Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "database error",
			filter: &domain.ScopeFilter{
				Path: func() *string { s := "/api/v1/users"; return &s }(),
			},
			mockSetup: func(m *MockScopeRepo) {
				m.On("Get", mock.Anything, mock.AnythingOfType("*domain.ScopeFilter")).
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
