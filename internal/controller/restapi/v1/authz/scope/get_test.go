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

func TestGet(t *testing.T) {
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
			query: "?path=/api/v1/users&method=GET",
			mockSetup: func(m *MockScopeUC) {
				m.On("Get", mock.Anything, mock.AnythingOfType("*domain.ScopeFilter")).
					Return(&domain.Scope{Path: "/api/v1/users", Method: "GET"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "bad_request_missing_path",
			query: "?method=GET",
			mockSetup: func(m *MockScopeUC) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "bad_request_missing_method",
			query: "?path=/api/v1/users",
			mockSetup: func(m *MockScopeUC) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "internal_error",
			query: "?path=/api/v1/users&method=GET",
			mockSetup: func(m *MockScopeUC) {
				m.On("Get", mock.Anything, mock.AnythingOfType("*domain.ScopeFilter")).
					Return(nil, errors.New("db error"))
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
			ctrl.Get(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockScope.AssertExpectations(t)
		})
	}
}
