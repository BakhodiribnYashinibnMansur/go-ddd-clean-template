package emailtemplate_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/emailtemplate"
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
			query: "?limit=10&offset=0&search=welcome",
			mockSetup: func(m *MockUseCase) {
				m.On("List", mock.Anything, mock.AnythingOfType("domain.EmailTemplateFilter")).
					Return([]domain.EmailTemplate{{Name: "welcome"}}, int64(1), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "success_default_pagination",
			query: "",
			mockSetup: func(m *MockUseCase) {
				m.On("List", mock.Anything, mock.AnythingOfType("domain.EmailTemplateFilter")).
					Return([]domain.EmailTemplate{}, int64(0), nil)
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
				m.On("List", mock.Anything, mock.AnythingOfType("domain.EmailTemplateFilter")).
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

			ctrl := emailtemplate.New(mockUC, cfg, log)
			ctrl.List(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
