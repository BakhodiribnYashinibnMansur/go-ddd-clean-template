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

func TestAddPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)
	validRoleID := uuid.New()
	validPermID := uuid.New()

	tests := []struct {
		name       string
		roleID     string
		permID     string
		mockSetup  func(*MockRoleUC)
		wantStatus int
	}{
		{
			name:   "success",
			roleID: validRoleID.String(),
			permID: validPermID.String(),
			mockSetup: func(m *MockRoleUC) {
				m.On("AddPermission", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "bad_request_invalid_role_id",
			roleID: "bad",
			permID: validPermID.String(),
			mockSetup: func(m *MockRoleUC) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "bad_request_invalid_perm_id",
			roleID: validRoleID.String(),
			permID: "bad",
			mockSetup: func(m *MockRoleUC) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "internal_error",
			roleID: validRoleID.String(),
			permID: validPermID.String(),
			mockSetup: func(m *MockRoleUC) {
				m.On("AddPermission", mock.Anything, mock.AnythingOfType("uuid.UUID"), mock.AnythingOfType("uuid.UUID")).
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
				{Key: "role_id", Value: tt.roleID},
				{Key: "perm_id", Value: tt.permID},
			}

			mockRole := new(MockRoleUC)
			tt.mockSetup(mockRole)

			ctrl := role.New(newTestUseCase(mockRole), cfg, log)
			ctrl.AddPermission(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRole.AssertExpectations(t)
		})
	}
}
