package policy_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/authz/policy"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)
	validID := uuid.New()

	tests := []struct {
		name       string
		paramID    string
		mockSetup  func(*MockPolicyUC)
		wantStatus int
	}{
		{
			name:    "success",
			paramID: validID.String(),
			mockSetup: func(m *MockPolicyUC) {
				m.On("Get", mock.Anything, mock.AnythingOfType("*domain.PolicyFilter")).
					Return(&domain.Policy{ID: validID, Effect: "allow"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "bad_request_invalid_uuid",
			paramID: "not-a-uuid",
			mockSetup: func(m *MockPolicyUC) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "internal_error",
			paramID: validID.String(),
			mockSetup: func(m *MockPolicyUC) {
				m.On("Get", mock.Anything, mock.AnythingOfType("*domain.PolicyFilter")).
					Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			c.Params = gin.Params{{Key: "policy_id", Value: tt.paramID}}

			mockPolicy := new(MockPolicyUC)
			tt.mockSetup(mockPolicy)

			ctrl := policy.New(newTestUseCase(mockPolicy), cfg, log)
			ctrl.Get(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockPolicy.AssertExpectations(t)
		})
	}
}
