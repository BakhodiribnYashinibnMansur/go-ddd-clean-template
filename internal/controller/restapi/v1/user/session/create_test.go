package session_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gct/internal/controller/restapi/v1/user/session"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		body       string
		mockSetup  func(*MockSessionUseCase)
		wantStatus int
	}{
		{
			name: "success",
			body: `{"user_id":"` + uuid.New().String() + `","device_type":"WEB"}`,
			mockSetup: func(m *MockSessionUseCase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Session")).
					Return(&domain.Session{ID: uuid.New()}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "bad_request_invalid_json",
			body:       `{invalid}`,
			mockSetup:  func(m *MockSessionUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "internal_error",
			body: `{"user_id":"` + uuid.New().String() + `","device_type":"WEB"}`,
			mockSetup: func(m *MockSessionUseCase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*domain.Session")).
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

			mockUC := new(MockSessionUseCase)
			tt.mockSetup(mockUC)

			ctrl := session.New(buildUseCase(mockUC), log)
			ctrl.Create(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
