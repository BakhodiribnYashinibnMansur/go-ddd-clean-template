package metric_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/controller/restapi/v1/audit/metric"
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
		mockSetup  func(*MockMetricUseCase)
		wantStatus int
	}{
		{
			name:  "success",
			query: "?limit=10&offset=0",
			mockSetup: func(m *MockMetricUseCase) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.FunctionMetricsFilter")).
					Return([]*domain.FunctionMetric{{ID: uuid.New(), Name: "TestFunc"}}, 1, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "bad_request_invalid_pagination",
			query:      "?limit=abc",
			mockSetup:  func(m *MockMetricUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "success_with_filters",
			query: "?limit=10&offset=0&name=TestFunc&is_panic=true",
			mockSetup: func(m *MockMetricUseCase) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.FunctionMetricsFilter")).
					Return([]*domain.FunctionMetric{}, 0, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "internal_error",
			query: "?limit=10&offset=0",
			mockSetup: func(m *MockMetricUseCase) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.FunctionMetricsFilter")).
					Return(nil, 0, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/metrics/functions"+tt.query, nil)

			mockUC := new(MockMetricUseCase)
			tt.mockSetup(mockUC)

			ctrl := metric.New(buildUseCase(mockUC), buildConfig(), log)
			ctrl.Gets(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
