package iprule_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/iprule"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		query      string
		mockSetup  func(*MockUseCase)
		wantStatus int
	}{
		{
			name:  "success",
			query: "?limit=10&offset=0&search=192&type=allow",
			mockSetup: func(m *MockUseCase) {
				m.On("List", mock.Anything, mock.AnythingOfType("domain.IPRuleFilter")).
					Return([]domain.IPRule{{IPAddress: "192.168.1.1"}}, int64(1), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "success_default_pagination",
			query: "",
			mockSetup: func(m *MockUseCase) {
				m.On("List", mock.Anything, mock.AnythingOfType("domain.IPRuleFilter")).
					Return([]domain.IPRule{}, int64(0), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "bad_request_invalid_limit",
			query: "?limit=abc",
			mockSetup: func(m *MockUseCase) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "internal_error",
			query: "?limit=10&offset=0",
			mockSetup: func(m *MockUseCase) {
				m.On("List", mock.Anything, mock.AnythingOfType("domain.IPRuleFilter")).
					Return(nil, int64(0), errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/"+tt.query, nil)

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := iprule.New(mockUC, cfg, log)
			ctrl.List(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
