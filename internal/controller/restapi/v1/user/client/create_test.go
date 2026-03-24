package client_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/user/client"
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
		mockSetup  func(*MockClientUseCase)
		wantStatus int
	}{
		{
			name: "success",
			body: `{"username":"testuser","phone":"+1234567890","password":"Secret123!"}`,
			mockSetup: func(m *MockClientUseCase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
					Return(nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "bad_request_invalid_json",
			body:       `{invalid}`,
			mockSetup:  func(m *MockClientUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "internal_error",
			body: `{"username":"testuser","phone":"+1234567890","password":"Secret123!"}`,
			mockSetup: func(m *MockClientUseCase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
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

			mockUC := new(MockClientUseCase)
			tt.mockSetup(mockUC)

			ctrl := client.New(buildUseCase(mockUC), cfg, log)
			ctrl.Create(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
