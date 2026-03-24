package setting_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi/v1/user/setting"
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
	validUID := uuid.New()

	tests := []struct {
		name       string
		userID     any
		setUserID  bool
		mockSetup  func(*MockUseCase)
		wantStatus int
	}{
		{
			name:      "success",
			userID:    validUID.String(),
			setUserID: true,
			mockSetup: func(m *MockUseCase) {
				m.On("Gets", mock.Anything, validUID).
					Return([]domain.UserSetting{{Key: "theme", Value: "dark"}}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "unauthorized_no_user_id",
			userID:     nil,
			setUserID:  false,
			mockSetup:  func(m *MockUseCase) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "bad_request_invalid_uuid",
			userID:     "not-a-uuid",
			setUserID:  true,
			mockSetup:  func(m *MockUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "internal_error",
			userID:    validUID.String(),
			setUserID: true,
			mockSetup: func(m *MockUseCase) {
				m.On("Gets", mock.Anything, validUID).
					Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			if tt.setUserID {
				c.Set(consts.CtxUserID, tt.userID)
			}

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := setting.New(mockUC, log)
			ctrl.Gets(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
