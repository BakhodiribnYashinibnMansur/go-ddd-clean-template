package dataexport_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/dataexport"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		id         string
		mockSetup  func(*MockUseCase)
		wantStatus int
	}{
		{
			name: "success",
			id:   "export-123",
			mockSetup: func(m *MockUseCase) {
				m.On("Delete", mock.Anything, "export-123").Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "bad_request_empty_id",
			id:         "",
			mockSetup:  func(m *MockUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "internal_error",
			id:   "export-123",
			mockSetup: func(m *MockUseCase) {
				m.On("Delete", mock.Anything, "export-123").Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)
			c.Params = gin.Params{
				{Key: "id", Value: tt.id},
			}

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := dataexport.New(mockUC, cfg, log)
			ctrl.Delete(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
