package dashboard_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/controller/restapi/v1/dashboard"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		mockSetup  func(*MockUseCase)
		wantStatus int
	}{
		{
			name: "success",
			mockSetup: func(m *MockUseCase) {
				m.On("Get", mock.Anything).
					Return(domain.DashboardStats{
						TotalUsers:     100,
						ActiveSessions: 25,
						TotalJobs:      10,
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "internal_error",
			mockSetup: func(m *MockUseCase) {
				m.On("Get", mock.Anything).
					Return(domain.DashboardStats{}, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := dashboard.New(mockUC, log)
			ctrl.Get(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
