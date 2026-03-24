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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChangeRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)
	validID := uuid.New()

	tests := []struct {
		name       string
		paramID    string
		body       string
		mockSetup  func(*MockClientUseCase)
		wantStatus int
	}{
		{
			name:    "success",
			paramID: validID.String(),
			body:    `{"role":"admin"}`,
			mockSetup: func(m *MockClientUseCase) {
				m.On("ChangeRole", mock.Anything, validID.String(), "admin").
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "bad_request_invalid_uuid",
			paramID:    "not-a-uuid",
			body:       `{"role":"admin"}`,
			mockSetup:  func(m *MockClientUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad_request_missing_role",
			paramID:    validID.String(),
			body:       `{}`,
			mockSetup:  func(m *MockClientUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad_request_invalid_json",
			paramID:    validID.String(),
			body:       `{invalid}`,
			mockSetup:  func(m *MockClientUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "internal_error",
			paramID: validID.String(),
			body:    `{"role":"admin"}`,
			mockSetup: func(m *MockClientUseCase) {
				m.On("ChangeRole", mock.Anything, validID.String(), "admin").
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
			c.Params = gin.Params{{Key: "user_id", Value: tt.paramID}}

			mockUC := new(MockClientUseCase)
			tt.mockSetup(mockUC)

			ctrl := client.New(buildUseCase(mockUC), cfg, log)
			ctrl.ChangeRole(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
