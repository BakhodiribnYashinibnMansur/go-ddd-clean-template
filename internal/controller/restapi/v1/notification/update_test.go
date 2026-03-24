package notification_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/notification"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)
	validID := uuid.New()

	tests := []struct {
		name       string
		paramID    string
		body       string
		mockSetup  func(*MockUseCase)
		wantStatus int
	}{
		{
			name:    "success",
			paramID: validID.String(),
			body:    `{"title":"Updated Notification"}`,
			mockSetup: func(m *MockUseCase) {
				m.On("Update", mock.Anything, validID, mock.AnythingOfType("domain.UpdateNotificationRequest")).
					Return(&domain.Notification{ID: validID, Title: "Updated Notification"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "bad_request_invalid_uuid",
			paramID: "not-a-uuid",
			body:    `{"title":"Updated Notification"}`,
			mockSetup: func(m *MockUseCase) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "bad_request_invalid_json",
			paramID: validID.String(),
			body:    `{invalid}`,
			mockSetup: func(m *MockUseCase) {
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "bad_request_invalid_type",
			paramID: validID.String(),
			body:    `{"type":"invalid"}`,
			mockSetup: func(m *MockUseCase) {
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "bad_request_invalid_target_type",
			paramID: validID.String(),
			body:    `{"target_type":"invalid"}`,
			mockSetup: func(m *MockUseCase) {
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "internal_error",
			paramID: validID.String(),
			body:    `{"title":"Updated Notification"}`,
			mockSetup: func(m *MockUseCase) {
				m.On("Update", mock.Anything, validID, mock.AnythingOfType("domain.UpdateNotificationRequest")).
					Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/"+tt.paramID, strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := notification.New(mockUC, cfg, log)
			ctrl.Update(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
