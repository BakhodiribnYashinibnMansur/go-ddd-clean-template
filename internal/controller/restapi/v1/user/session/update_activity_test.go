package session_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/controller/restapi/v1/user/session"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateActivity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New(logger.LevelInfo)

	validID := uuid.New().String()

	tests := []struct {
		name       string
		paramID    string
		mockSetup  func(*MockSessionUseCase)
		wantStatus int
	}{
		{
			name:    "success",
			paramID: validID,
			mockSetup: func(m *MockSessionUseCase) {
				m.On("UpdateActivity", mock.Anything, mock.AnythingOfType("*domain.SessionFilter")).
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "bad_request_invalid_id",
			paramID:    "not-a-uuid",
			mockSetup:  func(m *MockSessionUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "internal_error",
			paramID: validID,
			mockSetup: func(m *MockSessionUseCase) {
				m.On("UpdateActivity", mock.Anything, mock.AnythingOfType("*domain.SessionFilter")).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/sessions/"+tt.paramID+"/activity", nil)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}

			mockUC := new(MockSessionUseCase)
			tt.mockSetup(mockUC)

			ctrl := session.New(buildUseCase(mockUC), log)
			ctrl.UpdateActivity(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
