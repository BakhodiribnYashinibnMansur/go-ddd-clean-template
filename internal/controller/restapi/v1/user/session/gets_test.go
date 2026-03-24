package session_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/controller/restapi/v1/user/session"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSessions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		setUserID  bool
		query      string
		mockSetup  func(*MockSessionUseCase)
		wantStatus int
	}{
		{
			name:      "success",
			setUserID: true,
			query:     "?limit=10&offset=0",
			mockSetup: func(m *MockSessionUseCase) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.SessionsFilter")).
					Return([]*domain.Session{{ID: uuid.New()}}, 1, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "unauthorized_no_user_id",
			setUserID:  false,
			query:      "?limit=10&offset=0",
			mockSetup:  func(m *MockSessionUseCase) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:      "bad_request_invalid_pagination",
			setUserID: true,
			query:     "?limit=abc",
			mockSetup: func(m *MockSessionUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "internal_error",
			setUserID: true,
			query:     "?limit=10&offset=0",
			mockSetup: func(m *MockSessionUseCase) {
				m.On("Gets", mock.Anything, mock.AnythingOfType("*domain.SessionsFilter")).
					Return(nil, 0, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/sessions"+tt.query, nil)

			if tt.setUserID {
				c.Set("user_id", uuid.New())
			}

			mockUC := new(MockSessionUseCase)
			tt.mockSetup(mockUC)

			ctrl := session.New(buildUseCase(mockUC), log)
			ctrl.Sessions(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
