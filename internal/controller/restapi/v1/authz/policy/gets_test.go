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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		query      string
		mockSetup  func(*MockPolicyUC)
		wantStatus int
	}{
		{
			name:  "success",
			query: "?limit=10&offset=0",
			mockSetup: func(m *MockPolicyUC) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.PoliciesFilter")).
					Return([]*domain.Policy{{Effect: "allow"}}, 1, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "success_with_active_filter",
			query: "?limit=10&offset=0&active=true",
			mockSetup: func(m *MockPolicyUC) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.PoliciesFilter")).
					Return([]*domain.Policy{{Effect: "allow", Active: true}}, 1, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "bad_request_invalid_limit",
			query: "?limit=abc",
			mockSetup: func(m *MockPolicyUC) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "bad_request_limit_exceeds_max",
			query: "?limit=9999&offset=0",
			mockSetup: func(m *MockPolicyUC) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "internal_error",
			query: "?limit=10&offset=0",
			mockSetup: func(m *MockPolicyUC) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.PoliciesFilter")).
					Return(nil, 0, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/"+tt.query, nil)

			mockPolicy := new(MockPolicyUC)
			tt.mockSetup(mockPolicy)

			ctrl := policy.New(newTestUseCase(mockPolicy), cfg, log)
			ctrl.Gets(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockPolicy.AssertExpectations(t)
		})
	}
}
