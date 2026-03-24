package setting_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gct/internal/shared/domain/consts"
	"gct/internal/controller/restapi/v1/user/setting"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New(logger.LevelInfo)
	validUID := uuid.New()

	tests := []struct {
		name       string
		userID     any
		setUserID  bool
		key        string
		body       string
		mockSetup  func(*MockUseCase)
		wantStatus int
	}{
		{
			name:      "success_regular_key",
			userID:    validUID.String(),
			setUserID: true,
			key:       "theme",
			body:      `{"value":"dark"}`,
			mockSetup: func(m *MockUseCase) {
				m.On("Set", mock.Anything, validUID, "theme", "dark").Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "success_passcode_key",
			userID:    validUID.String(),
			setUserID: true,
			key:       "passcode",
			body:      `{"value":"1234"}`,
			mockSetup: func(m *MockUseCase) {
				m.On("SetPasscode", mock.Anything, validUID, "1234").Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "unauthorized_no_user_id",
			userID:     nil,
			setUserID:  false,
			key:        "theme",
			body:       `{"value":"dark"}`,
			mockSetup:  func(m *MockUseCase) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "bad_request_invalid_uuid",
			userID:     "not-a-uuid",
			setUserID:  true,
			key:        "theme",
			body:       `{"value":"dark"}`,
			mockSetup:  func(m *MockUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad_request_empty_key",
			userID:     validUID.String(),
			setUserID:  true,
			key:        "",
			body:       `{"value":"dark"}`,
			mockSetup:  func(m *MockUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad_request_invalid_body",
			userID:     validUID.String(),
			setUserID:  true,
			key:        "theme",
			body:       `{invalid}`,
			mockSetup:  func(m *MockUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad_request_missing_value",
			userID:     validUID.String(),
			setUserID:  true,
			key:        "theme",
			body:       `{}`,
			mockSetup:  func(m *MockUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "internal_error",
			userID:    validUID.String(),
			setUserID: true,
			key:       "theme",
			body:      `{"value":"dark"}`,
			mockSetup: func(m *MockUseCase) {
				m.On("Set", mock.Anything, validUID, "theme", "dark").
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{
				{Key: "key", Value: tt.key},
			}

			if tt.setUserID {
				c.Set(consts.CtxUserID, tt.userID)
			}

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := setting.New(mockUC, log)
			ctrl.Set(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
