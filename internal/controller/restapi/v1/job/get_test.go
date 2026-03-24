package job_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/job"
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
		mockSetup  func(*MockUseCase)
		wantStatus int
	}{
		{
			name:    "success",
			paramID: validID.String(),
			mockSetup: func(m *MockUseCase) {
				m.On("GetByID", mock.Anything, validID).
					Return(&domain.Job{ID: validID, Name: "test-job"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "bad_request_invalid_uuid",
			paramID: "not-a-uuid",
			mockSetup: func(m *MockUseCase) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "not_found",
			paramID: validID.String(),
			mockSetup: func(m *MockUseCase) {
				m.On("GetByID", mock.Anything, validID).
					Return(nil, errors.New("not found"))
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := job.New(mockUC, cfg, log)
			ctrl.Get(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
