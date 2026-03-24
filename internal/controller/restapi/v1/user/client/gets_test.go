package client_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/user/client"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		query      string
		mockSetup  func(*MockClientUseCase)
		wantStatus int
	}{
		{
			name:  "success",
			query: "?limit=10&offset=0",
			mockSetup: func(m *MockClientUseCase) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.UsersFilter")).
					Return([]*domain.User{{Username: func() *string { s := "test"; return &s }()}}, 1, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "success_default_pagination",
			query: "",
			mockSetup: func(m *MockClientUseCase) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.UsersFilter")).
					Return([]*domain.User{}, 0, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "bad_request_invalid_limit",
			query:      "?limit=abc",
			mockSetup:  func(m *MockClientUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "internal_error",
			query: "?limit=10&offset=0",
			mockSetup: func(m *MockClientUseCase) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.UsersFilter")).
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

			mockUC := new(MockClientUseCase)
			tt.mockSetup(mockUC)

			ctrl := client.New(buildUseCase(mockUC), cfg, log)
			ctrl.Users(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
