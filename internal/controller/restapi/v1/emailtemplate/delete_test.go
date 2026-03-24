package emailtemplate_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/emailtemplate"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		paramID    string
		mockSetup  func(*MockUseCase)
		wantStatus int
	}{
		{
			name:    "success",
			paramID: "welcome-template",
			mockSetup: func(m *MockUseCase) {
				m.On("Delete", mock.Anything, "welcome-template").Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "internal_error",
			paramID: "welcome-template",
			mockSetup: func(m *MockUseCase) {
				m.On("Delete", mock.Anything, "welcome-template").Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := emailtemplate.New(mockUC, cfg, log)
			ctrl.Delete(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
