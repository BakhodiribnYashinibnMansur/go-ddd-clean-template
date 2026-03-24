package scope_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/authz/scope"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		query      string
		mockSetup  func(*MockScopeUC)
		wantStatus int
	}{
		{
			name:  "success",
			query: "?limit=10&offset=0",
			mockSetup: func(m *MockScopeUC) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.ScopesFilter")).
					Return([]*domain.Scope{{Path: "/api/v1/users", Method: "GET"}}, 1, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "success_with_filters",
			query: "?limit=10&offset=0&path=/api/v1/users&method=GET",
			mockSetup: func(m *MockScopeUC) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.ScopesFilter")).
					Return([]*domain.Scope{{Path: "/api/v1/users", Method: "GET"}}, 1, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "bad_request_invalid_limit",
			query: "?limit=abc&offset=0",
			mockSetup: func(m *MockScopeUC) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "bad_request_limit_exceeds_max",
			query: "?limit=9999&offset=0",
			mockSetup: func(m *MockScopeUC) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "internal_error",
			query: "?limit=10&offset=0",
			mockSetup: func(m *MockScopeUC) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.ScopesFilter")).
					Return(nil, 0, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/"+tt.query, nil)

			mockScope := new(MockScopeUC)
			tt.mockSetup(mockScope)

			ctrl := scope.New(newTestUseCase(mockScope), cfg, log)
			ctrl.Gets(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockScope.AssertExpectations(t)
		})
	}
}
