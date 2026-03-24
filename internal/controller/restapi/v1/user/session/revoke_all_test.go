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

func TestRevokeAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name         string
		setUserID    bool
		setSessionID bool
		mockSetup    func(*MockSessionUseCase)
		wantStatus   int
	}{
		{
			name:         "success",
			setUserID:    true,
			setSessionID: true,
			mockSetup: func(m *MockSessionUseCase) {
				m.On("Revoke", mock.Anything, mock.AnythingOfType("*domain.SessionFilter")).
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:         "unauthorized_no_user_id",
			setUserID:    false,
			setSessionID: true,
			mockSetup:    func(m *MockSessionUseCase) {},
			wantStatus:   http.StatusUnauthorized,
		},
		{
			name:         "unauthorized_no_session_id",
			setUserID:    true,
			setSessionID: false,
			mockSetup:    func(m *MockSessionUseCase) {},
			wantStatus:   http.StatusUnauthorized,
		},
		{
			name:         "internal_error",
			setUserID:    true,
			setSessionID: true,
			mockSetup: func(m *MockSessionUseCase) {
				m.On("Revoke", mock.Anything, mock.AnythingOfType("*domain.SessionFilter")).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/sessions/revoke-all", nil)
			c.Request.Header.Set("Content-Type", "application/json")

			if tt.setUserID {
				c.Set("user_id", uuid.New())
			}
			if tt.setSessionID {
				c.Set("session_id", uuid.New())
			}

			mockUC := new(MockSessionUseCase)
			tt.mockSetup(mockUC)

			ctrl := session.New(buildUseCase(mockUC), log)
			ctrl.RevokeAll(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
