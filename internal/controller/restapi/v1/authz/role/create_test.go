package role_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/authz/role"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		body       string
		mockSetup  func(*MockRoleUC)
		wantStatus int
	}{
		{
			name: "success",
			body: `{"name":"admin"}`,
			mockSetup: func(m *MockRoleUC) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Role")).Return(nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "bad_request_invalid_json",
			body: `{invalid`,
			mockSetup: func(m *MockRoleUC) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "internal_error",
			body: `{"name":"admin"}`,
			mockSetup: func(m *MockRoleUC) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Role")).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			mockRole := new(MockRoleUC)
			tt.mockSetup(mockRole)

			ctrl := role.New(newTestUseCase(mockRole), cfg, log)
			ctrl.Create(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRole.AssertExpectations(t)
		})
	}
}
