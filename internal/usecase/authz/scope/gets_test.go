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

func TestUseCase_Gets(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name       string
		filter     *domain.ScopesFilter
		mockSetup  func(*MockScopeRepo)
		wantErr    bool
		wantCount  int
		wantScopes int
	}{
		{
			name: "success - multiple results",
			filter: &domain.ScopesFilter{
				ScopeFilter: domain.ScopeFilter{},
			},
			mockSetup: func(m *MockScopeRepo) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.ScopesFilter")).
					Return([]*domain.Scope{
						{Path: "/api/v1/users", Method: "GET", CreatedAt: now},
						{Path: "/api/v1/users", Method: "POST", CreatedAt: now},
						{Path: "/api/v1/admin", Method: "GET", CreatedAt: now},
					}, 3, nil)
			},
			wantErr:    false,
			wantCount:  3,
			wantScopes: 3,
		},
		{
			name: "success - with filter",
			filter: &domain.ScopesFilter{
				ScopeFilter: domain.ScopeFilter{
					Path: func() *string { s := "/api/v1/users"; return &s }(),
				},
			},
			mockSetup: func(m *MockScopeRepo) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.ScopesFilter")).
					Return([]*domain.Scope{
						{Path: "/api/v1/users", Method: "GET", CreatedAt: now},
						{Path: "/api/v1/users", Method: "POST", CreatedAt: now},
					}, 2, nil)
			},
			wantErr:    false,
			wantCount:  2,
			wantScopes: 2,
		},
		{
			name: "success - with pagination",
			filter: &domain.ScopesFilter{
				ScopeFilter: domain.ScopeFilter{},
				Pagination:  &domain.Pagination{Limit: 10, Offset: 0},
			},
			mockSetup: func(m *MockScopeRepo) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.ScopesFilter")).
					Return([]*domain.Scope{
						{Path: "/api/v1/users", Method: "GET", CreatedAt: now},
					}, 50, nil)
			},
			wantErr:    false,
			wantCount:  50,
			wantScopes: 1,
		},
		{
			name: "success - empty result",
			filter: &domain.ScopesFilter{
				ScopeFilter: domain.ScopeFilter{},
			},
			mockSetup: func(m *MockScopeRepo) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.ScopesFilter")).
					Return([]*domain.Scope{}, 0, nil)
			},
			wantErr:    false,
			wantCount:  0,
			wantScopes: 0,
		},
		{
			name: "repo error",
			filter: &domain.ScopesFilter{
				ScopeFilter: domain.ScopeFilter{},
			},
			mockSetup: func(m *MockScopeRepo) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.ScopesFilter")).
					Return([]*domain.Scope(nil), 0, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, repo := setup(t)
			tt.mockSetup(repo)

			scopes, count, err := uc.Gets(t.Context(), tt.filter)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, scopes)
				assert.Equal(t, 0, count)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCount, count)
				assert.Len(t, scopes, tt.wantScopes)
			}

			repo.AssertExpectations(t)
		})
	}
}
