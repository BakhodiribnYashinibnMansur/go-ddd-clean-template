package role_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/authz/role"
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
		mockSetup  func(*MockRoleUC)
		wantStatus int
	}{
		{
			name:  "success",
			query: "?limit=10&offset=0",
			mockSetup: func(m *MockRoleUC) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.RolesFilter")).
					Return([]*domain.Role{{Name: "admin"}}, 1, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "bad_request_missing_pagination",
			query: "",
			mockSetup: func(m *MockRoleUC) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.RolesFilter")).
					Return([]*domain.Role{}, 0, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "internal_error",
			query: "?limit=10&offset=0",
			mockSetup: func(m *MockRoleUC) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.RolesFilter")).
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

			mockRole := new(MockRoleUC)
			tt.mockSetup(mockRole)

			ctrl := role.New(newTestUseCase(mockRole), cfg, log)
			ctrl.Gets(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRole.AssertExpectations(t)
		})
	}
}
