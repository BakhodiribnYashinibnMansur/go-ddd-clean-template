package scope_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUseCase_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		path      string
		method    string
		mockSetup func(*MockScopeRepo)
		wantErr   bool
	}{
		{
			name:   "success",
			path:   "/api/v1/users",
			method: "GET",
			mockSetup: func(m *MockScopeRepo) {
				m.On("Delete", mock.Anything, "/api/v1/users", "GET").Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "success - DELETE method",
			path:   "/api/v1/users/:id",
			method: "DELETE",
			mockSetup: func(m *MockScopeRepo) {
				m.On("Delete", mock.Anything, "/api/v1/users/:id", "DELETE").Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "not found",
			path:   "/api/v1/nonexistent",
			method: "GET",
			mockSetup: func(m *MockScopeRepo) {
				m.On("Delete", mock.Anything, "/api/v1/nonexistent", "GET").
					Return(errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name:   "database error",
			path:   "/api/v1/users",
			method: "GET",
			mockSetup: func(m *MockScopeRepo) {
				m.On("Delete", mock.Anything, "/api/v1/users", "GET").
					Return(errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name:   "empty path and method",
			path:   "",
			method: "",
			mockSetup: func(m *MockScopeRepo) {
				m.On("Delete", mock.Anything, "", "").
					Return(errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc, repo := setup(t)
			tt.mockSetup(repo)

			err := uc.Delete(t.Context(), tt.path, tt.method)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}
