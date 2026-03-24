package dataexport_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/dataexport"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		body       string
		userID     string
		mockSetup  func(*MockUseCase)
		wantStatus int
	}{
		{
			name:   "success",
			body:   `{"type":"users","filters":{"status":"active"}}`,
			userID: "user-123",
			mockSetup: func(m *MockUseCase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.CreateDataExportRequest"), "user-123").
					Return(&domain.DataExport{ID: "export-1", Type: "users", Status: "pending"}, nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "bad_request_invalid_json",
			body:       `{invalid}`,
			userID:     "user-123",
			mockSetup:  func(m *MockUseCase) {},
			wantStatus: http.StatusOK, // httpx.BindJSON returns false without setting status (middleware handles it)
		},
		{
			name:   "internal_error",
			body:   `{"type":"users"}`,
			userID: "user-123",
			mockSetup: func(m *MockUseCase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.CreateDataExportRequest"), "user-123").
					Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:   "success_no_user_id",
			body:   `{"type":"users"}`,
			userID: "",
			mockSetup: func(m *MockUseCase) {
				m.On("Create", mock.Anything, mock.AnythingOfType("domain.CreateDataExportRequest"), "").
					Return(&domain.DataExport{ID: "export-1", Type: "users", Status: "pending"}, nil)
			},
			wantStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := dataexport.New(mockUC, cfg, log)
			ctrl.Create(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
