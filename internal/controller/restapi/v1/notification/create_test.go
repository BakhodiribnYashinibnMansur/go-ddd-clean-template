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
		mockSetup  func(*MockUseCase)
		wantStatus int
	}{
		{
			name: "success",
			body: `{"title":"System Update","body":"Scheduled maintenance tonight","type":"info","target_type":"all","is_active":true}`,
			mockSetup: func(m *MockUseCase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.CreateNotificationRequest")).
					Return(&domain.Notification{Title: "System Update"}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "bad_request_missing_title",
			body: `{"body":"Some body","type":"info","target_type":"all"}`,
			mockSetup: func(m *MockUseCase) {
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "bad_request_missing_body",
			body: `{"title":"Test","type":"info","target_type":"all"}`,
			mockSetup: func(m *MockUseCase) {
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "bad_request_invalid_type",
			body: `{"title":"Test","body":"Body","type":"invalid","target_type":"all"}`,
			mockSetup: func(m *MockUseCase) {
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "bad_request_invalid_target_type",
			body: `{"title":"Test","body":"Body","type":"info","target_type":"invalid"}`,
			mockSetup: func(m *MockUseCase) {
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "bad_request_invalid_json",
			body: `{invalid}`,
			mockSetup: func(m *MockUseCase) {
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "internal_error",
			body: `{"title":"System Update","body":"Scheduled maintenance","type":"info","target_type":"all"}`,
			mockSetup: func(m *MockUseCase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.CreateNotificationRequest")).
					Return(nil, errors.New("db error"))
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

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := notification.New(mockUC, cfg, log)
			ctrl.Create(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
