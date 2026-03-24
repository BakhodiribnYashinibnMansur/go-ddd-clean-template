package role_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/authz/role"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAssign(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)
	validUserID := uuid.New()
	validRoleID := uuid.New()

	tests := []struct {
		name       string
		userID     string
		roleID     string
		mockSetup  func(*MockRoleUC)
		wantStatus int
	}{
		{
			name:   "success",
			userID: validUserID.String(),
			roleID: validRoleID.String(),
			mockSetup: func(m *MockRoleUC) {
				m.On("Assign", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "bad_request_invalid_user_id",
			userID: "bad",
			roleID: validRoleID.String(),
			mockSetup: func(m *MockRoleUC) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "bad_request_invalid_role_id",
			userID: validUserID.String(),
			roleID: "bad",
			mockSetup: func(m *MockRoleUC) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "internal_error",
			userID: validUserID.String(),
			roleID: validRoleID.String(),
			mockSetup: func(m *MockRoleUC) {
				m.On("Assign", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
			c.Params = gin.Params{
				{Key: "user_id", Value: tt.userID},
				{Key: "role_id", Value: tt.roleID},
			}

			mockRole := new(MockRoleUC)
			tt.mockSetup(mockRole)

			ctrl := role.New(newTestUseCase(mockRole), cfg, log)
			ctrl.Assign(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRole.AssertExpectations(t)
		})
	}
}
