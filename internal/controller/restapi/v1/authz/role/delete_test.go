package role_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/authz/role"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)
	validID := uuid.New()

	tests := []struct {
		name       string
		paramID    string
		mockSetup  func(*MockRoleUC)
		wantStatus int
	}{
		{
			name:    "success",
			paramID: validID.String(),
			mockSetup: func(m *MockRoleUC) {
				m.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "bad_request_invalid_uuid",
			paramID: "bad",
			mockSetup: func(m *MockRoleUC) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "internal_error",
			paramID: validID.String(),
			mockSetup: func(m *MockRoleUC) {
				m.On("Delete", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)
			c.Params = gin.Params{{Key: "role_id", Value: tt.paramID}}

			mockRole := new(MockRoleUC)
			tt.mockSetup(mockRole)

			ctrl := role.New(newTestUseCase(mockRole), cfg, log)
			ctrl.Delete(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRole.AssertExpectations(t)
		})
	}
}
