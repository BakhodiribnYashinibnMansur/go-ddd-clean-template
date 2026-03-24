package permission_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/authz/permission"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)
	validID := uuid.New()

	tests := []struct {
		name       string
		paramID    string
		mockSetup  func(*MockPermUC)
		wantStatus int
	}{
		{
			name:    "success",
			paramID: validID.String(),
			mockSetup: func(m *MockPermUC) {
				m.On("Get", mock.Anything, mock.AnythingOfType("*domain.PermissionFilter")).
					Return(&domain.Permission{ID: validID, Name: "read_users"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "bad_request_invalid_uuid",
			paramID:    "not-a-uuid",
			mockSetup:  func(m *MockPermUC) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "not_found_nil_result",
			paramID: validID.String(),
			mockSetup: func(m *MockPermUC) {
				m.On("Get", mock.Anything, mock.AnythingOfType("*domain.PermissionFilter")).
					Return(nil, errors.New("not found"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:    "internal_error",
			paramID: validID.String(),
			mockSetup: func(m *MockPermUC) {
				m.On("Get", mock.Anything, mock.AnythingOfType("*domain.PermissionFilter")).
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
			c.Params = gin.Params{{Key: "perm_id", Value: tt.paramID}}

			mockPerm := new(MockPermUC)
			tt.mockSetup(mockPerm)

			ctrl := permission.New(newTestUseCase(mockPerm), cfg, log)
			ctrl.Get(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockPerm.AssertExpectations(t)
		})
	}
}
