package job_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/job"
	"gct/internal/domain"
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
		mockSetup  func(*MockUseCase)
		wantStatus int
	}{
		{
			name: "success",
			body: `{"name":"cleanup-job","type":"cron","cron_schedule":"0 * * * *"}`,
			mockSetup: func(m *MockUseCase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.CreateJobRequest")).
					Return(&domain.Job{Name: "cleanup-job"}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "bad_request_missing_name",
			body: `{"type":"cron"}`,
			mockSetup: func(m *MockUseCase) {
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "bad_request_invalid_json",
			body: `{invalid}`,
			mockSetup: func(m *MockUseCase) {
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "internal_error",
			body: `{"name":"cleanup-job","type":"cron","cron_schedule":"0 * * * *"}`,
			mockSetup: func(m *MockUseCase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.CreateJobRequest")).
					Return(nil, errors.New("db error"))
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

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := job.New(mockUC, cfg, log)
			ctrl.Create(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
