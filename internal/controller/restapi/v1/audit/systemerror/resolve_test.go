package systemerror_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	auditsystemerror "gct/internal/controller/restapi/v1/audit/systemerror"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResolve(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New(logger.LevelInfo)

	validID := uuid.New().String()

	tests := []struct {
		name       string
		paramID    string
		setUserID  bool
		mockSetup  func(*MockSystemErrorUseCase)
		wantStatus int
	}{
		{
			name:      "success_with_user_id",
			paramID:   validID,
			setUserID: true,
			mockSetup: func(m *MockSystemErrorUseCase) {
				m.On("Resolve", mock.Anything, validID, mock.AnythingOfType("*string")).
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "success_without_user_id",
			paramID:   validID,
			setUserID: false,
			mockSetup: func(m *MockSystemErrorUseCase) {
				m.On("Resolve", mock.Anything, validID, (*string)(nil)).
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "bad_request_invalid_id",
			paramID:    "not-a-uuid",
			setUserID:  false,
			mockSetup:  func(m *MockSystemErrorUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "internal_error",
			paramID:   validID,
			setUserID: true,
			mockSetup: func(m *MockSystemErrorUseCase) {
				m.On("Resolve", mock.Anything, validID, mock.AnythingOfType("*string")).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/system/errors/"+tt.paramID+"/resolve", nil)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}

			if tt.setUserID {
				c.Set("user_id", uuid.New())
			}

			mockUC := new(MockSystemErrorUseCase)
			tt.mockSetup(mockUC)

			ctrl := auditsystemerror.New(mockUC, log)
			ctrl.Resolve(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
