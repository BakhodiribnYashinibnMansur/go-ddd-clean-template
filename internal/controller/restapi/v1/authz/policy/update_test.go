package policy_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/authz/policy"
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
		mockSetup  func(*MockPolicyUC)
		wantStatus int
	}{
		{
			name:    "success",
			paramID: validID.String(),
			body:    `{"effect":"deny","priority":2,"active":true,"conditions":{}}`,
			mockSetup: func(m *MockPolicyUC) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.Policy")).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "bad_request_invalid_uuid",
			paramID: "not-a-uuid",
			body:    `{"effect":"deny"}`,
			mockSetup: func(m *MockPolicyUC) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "bad_request_invalid_json",
			paramID: validID.String(),
			body:    `{invalid`,
			mockSetup: func(m *MockPolicyUC) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "internal_error",
			paramID: validID.String(),
			body:    `{"effect":"deny","priority":2,"active":true,"conditions":{}}`,
			mockSetup: func(m *MockPolicyUC) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.Policy")).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "policy_id", Value: tt.paramID}}

			mockPolicy := new(MockPolicyUC)
			tt.mockSetup(mockPolicy)

			ctrl := policy.New(newTestUseCase(mockPolicy), cfg, log)
			ctrl.Update(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockPolicy.AssertExpectations(t)
		})
	}
}
