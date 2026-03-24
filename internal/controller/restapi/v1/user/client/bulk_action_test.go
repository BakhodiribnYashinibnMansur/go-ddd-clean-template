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

func TestBulkAction(t *testing.T) {
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
			body: `{"ids":["id-1","id-2"],"action":"deactivate"}`,
			mockSetup: func(m *MockClientUseCase) {
				m.On("BulkAction", mock.Anything, mock.AnythingOfType("domain.BulkActionRequest")).
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "bad_request_invalid_json",
			body:       `{invalid}`,
			mockSetup:  func(m *MockClientUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad_request_missing_required",
			body:       `{"ids":[],"action":""}`,
			mockSetup:  func(m *MockClientUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "internal_error",
			body: `{"ids":["id-1"],"action":"delete"}`,
			mockSetup: func(m *MockClientUseCase) {
				m.On("BulkAction", mock.Anything, mock.AnythingOfType("domain.BulkActionRequest")).
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
			ctrl.BulkAction(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
