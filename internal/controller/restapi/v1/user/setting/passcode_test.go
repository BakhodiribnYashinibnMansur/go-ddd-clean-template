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

func TestVerifyPasscode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New(logger.LevelInfo)
	validUID := uuid.New()

	tests := []struct {
		name       string
		userID     any
		setUserID  bool
		body       string
		mockSetup  func(*MockUseCase)
		wantStatus int
	}{
		{
			name:      "success",
			userID:    validUID.String(),
			setUserID: true,
			body:      `{"passcode":"1234"}`,
			mockSetup: func(m *MockUseCase) {
				m.On("VerifyPasscode", mock.Anything, validUID, "1234").Return(true, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "incorrect_passcode",
			userID:    validUID.String(),
			setUserID: true,
			body:      `{"passcode":"0000"}`,
			mockSetup: func(m *MockUseCase) {
				m.On("VerifyPasscode", mock.Anything, validUID, "0000").Return(false, nil)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "unauthorized_no_user_id",
			userID:     nil,
			setUserID:  false,
			body:       `{"passcode":"1234"}`,
			mockSetup:  func(m *MockUseCase) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "bad_request_invalid_uuid",
			userID:     "not-a-uuid",
			setUserID:  true,
			body:       `{"passcode":"1234"}`,
			mockSetup:  func(m *MockUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad_request_invalid_body",
			userID:     validUID.String(),
			setUserID:  true,
			body:       `{invalid}`,
			mockSetup:  func(m *MockUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad_request_missing_passcode",
			userID:     validUID.String(),
			setUserID:  true,
			body:       `{}`,
			mockSetup:  func(m *MockUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "internal_error",
			userID:    validUID.String(),
			setUserID: true,
			body:      `{"passcode":"1234"}`,
			mockSetup: func(m *MockUseCase) {
				m.On("VerifyPasscode", mock.Anything, validUID, "1234").
					Return(false, errors.New("db error"))
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

			if tt.setUserID {
				c.Set(consts.CtxUserID, tt.userID)
			}

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := setting.New(mockUC, log)
			ctrl.VerifyPasscode(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}

func TestRemovePasscode(t *testing.T) {
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
				m.On("RemovePasscode", mock.Anything, validUID).Return(nil)
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
				m.On("RemovePasscode", mock.Anything, validUID).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)

			if tt.setUserID {
				c.Set(consts.CtxUserID, tt.userID)
			}

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := setting.New(mockUC, log)
			ctrl.RemovePasscode(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
