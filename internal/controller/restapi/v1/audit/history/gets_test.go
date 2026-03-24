package history_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/controller/restapi/v1/audit/history"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		query      string
		mockSetup  func(*MockEndpointHistoryUseCase)
		wantStatus int
	}{
		{
			name:  "success",
			query: "?limit=10&offset=0",
			mockSetup: func(m *MockEndpointHistoryUseCase) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.EndpointHistoriesFilter")).
					Return([]*domain.EndpointHistory{{ID: uuid.New()}}, 1, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "bad_request_invalid_pagination",
			query:      "?limit=abc",
			mockSetup:  func(m *MockEndpointHistoryUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "success_with_filters",
			query: "?limit=10&offset=0&method=GET&path=/api/v1/users&status_code=200",
			mockSetup: func(m *MockEndpointHistoryUseCase) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.EndpointHistoriesFilter")).
					Return([]*domain.EndpointHistory{}, 0, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "internal_error",
			query: "?limit=10&offset=0",
			mockSetup: func(m *MockEndpointHistoryUseCase) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.EndpointHistoriesFilter")).
					Return(nil, 0, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/audit/history"+tt.query, nil)

			mockUC := new(MockEndpointHistoryUseCase)
			tt.mockSetup(mockUC)

			ctrl := history.New(buildUseCase(mockUC), buildConfig(), log)
			ctrl.Gets(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
