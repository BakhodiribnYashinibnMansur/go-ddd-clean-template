package errorcode_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gct/internal/controller/restapi/v1/errorcode"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		paramCode  string
		body       string
		mockSetup  func(*MockUseCase)
		wantStatus int
	}{
		{
			name:      "success",
			paramCode: "ERR_001",
			body:      `{"message":"Updated message"}`,
			mockSetup: func(m *MockUseCase) {
				m.On("Update", mock.Anything, "ERR_001", mock.Anything).
					Return(&domain.ErrorCode{Code: "ERR_001", Message: "Updated message"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "bad_request_empty_code",
			paramCode: "",
			body:      `{"message":"Updated message"}`,
			mockSetup: func(m *MockUseCase) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "bad_request_invalid_json",
			paramCode: "ERR_001",
			body:      `{invalid}`,
			mockSetup: func(m *MockUseCase) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "internal_error",
			paramCode: "ERR_001",
			body:      `{"message":"Updated message"}`,
			mockSetup: func(m *MockUseCase) {
				m.On("Update", mock.Anything, "ERR_001", mock.Anything).
					Return(nil, errors.New("db error"))
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
			c.Params = gin.Params{{Key: "code", Value: tt.paramCode}}

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := errorcode.New(mockUC, log)
			ctrl.Update(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
