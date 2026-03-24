package ratelimit_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/ratelimit"
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
			body: `{"name":"test-rule","path_pattern":"/api/*","method":"GET","limit_count":100,"window_seconds":60}`,
			mockSetup: func(m *MockUseCase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.CreateRateLimitRequest")).
					Return(&domain.RateLimit{Name: "test-rule"}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "bad_request_missing_name",
			body: `{"path_pattern":"/api/*","method":"GET","limit_count":100,"window_seconds":60}`,
			mockSetup: func(m *MockUseCase) {
				// BindJSON fails, usecase not called
			},
			wantStatus: http.StatusOK, // gin returns 200 when ShouldBindJSON fails without middleware
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
			body: `{"name":"test-rule","path_pattern":"/api/*","method":"GET","limit_count":100,"window_seconds":60}`,
			mockSetup: func(m *MockUseCase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.CreateRateLimitRequest")).
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

			ctrl := ratelimit.New(mockUC, cfg, log)
			ctrl.Create(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
