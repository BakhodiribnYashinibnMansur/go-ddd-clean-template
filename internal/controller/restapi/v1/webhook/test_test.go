package webhook_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/webhook"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTestWebhook(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)
	validID := uuid.New()

	tests := []struct {
		name       string
		paramID    string
		mockSetup  func(*MockUseCase)
		wantStatus int
	}{
		{
			name:    "success",
			paramID: validID.String(),
			mockSetup: func(m *MockUseCase) {
				m.On("Test", mock.Anything, validID).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "bad_request_invalid_uuid",
			paramID: "not-a-uuid",
			mockSetup: func(m *MockUseCase) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "bad_request_empty_id",
			paramID: "",
			mockSetup: func(m *MockUseCase) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "internal_error",
			paramID: validID.String(),
			mockSetup: func(m *MockUseCase) {
				m.On("Test", mock.Anything, validID).Return(errors.New("webhook delivery failed"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/"+tt.paramID+"/test", nil)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := webhook.New(mockUC, cfg, log)
			ctrl.Test(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
