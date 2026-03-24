package permission_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/authz/permission"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRemoveScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)
	validID := uuid.New()

	tests := []struct {
		name       string
		paramID    string
		body       string
		mockSetup  func(*MockPermUC)
		wantStatus int
	}{
		{
			name:    "success",
			paramID: validID.String(),
			body:    `{"path":"/api/v1/users","method":"GET"}`,
			mockSetup: func(m *MockPermUC) {
				m.On("RemoveScope", mock.Anything, mock.AnythingOfType("uuid.UUID"), "/api/v1/users", "GET").
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "bad_request_invalid_uuid",
			paramID:    "bad",
			body:       `{"path":"/api/v1/users","method":"GET"}`,
			mockSetup:  func(m *MockPermUC) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad_request_invalid_json",
			paramID:    validID.String(),
			body:       `{invalid`,
			mockSetup:  func(m *MockPermUC) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad_request_missing_path",
			paramID:    validID.String(),
			body:       `{"method":"DELETE"}`,
			mockSetup:  func(m *MockPermUC) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad_request_missing_method",
			paramID:    validID.String(),
			body:       `{"path":"/api/v1/users"}`,
			mockSetup:  func(m *MockPermUC) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "internal_error",
			paramID: validID.String(),
			body:    `{"path":"/api/v1/users","method":"GET"}`,
			mockSetup: func(m *MockPermUC) {
				m.On("RemoveScope", mock.Anything, mock.AnythingOfType("uuid.UUID"), "/api/v1/users", "GET").
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "perm_id", Value: tt.paramID}}

			mockPerm := new(MockPermUC)
			tt.mockSetup(mockPerm)

			ctrl := permission.New(newTestUseCase(mockPerm), cfg, log)
			ctrl.RemoveScope(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockPerm.AssertExpectations(t)
		})
	}
}
