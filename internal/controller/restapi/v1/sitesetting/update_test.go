package sitesetting_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/sitesetting"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateByKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		key        string
		body       string
		mockSetup  func(*MockUseCase)
		wantStatus int
	}{
		{
			name: "success",
			key:  "site_name",
			body: `{"value":"New Site Name"}`,
			mockSetup: func(m *MockUseCase) {
				m.On("UpdateByKey", mock.Anything, "site_name", "New Site Name").
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "bad_request_empty_key",
			key:        "",
			body:       `{"value":"New Site Name"}`,
			mockSetup:  func(m *MockUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad_request_invalid_body",
			key:        "site_name",
			body:       `{invalid}`,
			mockSetup:  func(m *MockUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad_request_missing_value",
			key:        "site_name",
			body:       `{}`,
			mockSetup:  func(m *MockUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "internal_error",
			key:  "site_name",
			body: `{"value":"New Site Name"}`,
			mockSetup: func(m *MockUseCase) {
				m.On("UpdateByKey", mock.Anything, "site_name", "New Site Name").
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

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := sitesetting.New(mockUC, cfg, log)
			ctrl.UpdateByKey(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
