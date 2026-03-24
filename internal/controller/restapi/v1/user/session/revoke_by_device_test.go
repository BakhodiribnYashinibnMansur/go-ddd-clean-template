package session_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/controller/restapi/v1/user/session"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRevokeByDevice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		setUserID  bool
		deviceID   string
		wantStatus int
	}{
		{
			name:       "success",
			setUserID:  true,
			deviceID:   "device-abc-123",
			wantStatus: http.StatusOK,
		},
		{
			name:       "unauthorized_no_user_id",
			setUserID:  false,
			deviceID:   "device-abc-123",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "bad_request_empty_device_id",
			setUserID:  true,
			deviceID:   "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/sessions/device/"+tt.deviceID, nil)
			c.Params = gin.Params{{Key: "device_id", Value: tt.deviceID}}

			if tt.setUserID {
				c.Set("user_id", uuid.New())
			}

			mockUC := new(MockSessionUseCase)

			ctrl := session.New(buildUseCase(mockUC), log)
			ctrl.RevokeByDevice(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
