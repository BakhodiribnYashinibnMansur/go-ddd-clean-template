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
		mockSetup  func(*MockPolicyUC)
		wantStatus int
	}{
		{
			name: "success",
			body: `{"permission_id":"00000000-0000-0000-0000-000000000001","effect":"allow","priority":1,"active":true,"conditions":{}}`,
			mockSetup: func(m *MockPolicyUC) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Policy")).Return(nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "bad_request_invalid_json",
			body: `{invalid`,
			mockSetup: func(m *MockPolicyUC) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "internal_error",
			body: `{"permission_id":"00000000-0000-0000-0000-000000000001","effect":"allow","priority":1,"active":true,"conditions":{}}`,
			mockSetup: func(m *MockPolicyUC) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Policy")).
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

			mockPolicy := new(MockPolicyUC)
			tt.mockSetup(mockPolicy)

			ctrl := policy.New(newTestUseCase(mockPolicy), cfg, log)
			ctrl.Create(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockPolicy.AssertExpectations(t)
		})
	}
}
