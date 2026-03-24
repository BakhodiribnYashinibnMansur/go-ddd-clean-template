package permission_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/authz/permission"
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
		mockSetup  func(*MockPermUC)
		wantStatus int
	}{
		{
			name:  "success",
			query: "?limit=10&offset=0",
			mockSetup: func(m *MockPermUC) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.PermissionsFilter")).
					Return([]*domain.Permission{{Name: "read_users"}}, 1, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "success_with_name_filter",
			query: "?limit=10&offset=0&name=read",
			mockSetup: func(m *MockPermUC) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.PermissionsFilter")).
					Return([]*domain.Permission{{Name: "read_users"}}, 1, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "bad_request_invalid_limit",
			query:      "?limit=abc&offset=0",
			mockSetup:  func(m *MockPermUC) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "internal_error",
			query: "?limit=10&offset=0",
			mockSetup: func(m *MockPermUC) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.PermissionsFilter")).
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

			mockPerm := new(MockPermUC)
			tt.mockSetup(mockPerm)

			ctrl := permission.New(newTestUseCase(mockPerm), cfg, log)
			ctrl.Gets(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockPerm.AssertExpectations(t)
		})
	}
}
