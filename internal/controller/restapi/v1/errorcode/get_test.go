package errorcode_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/controller/restapi/v1/errorcode"
	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		paramCode  string
		mockSetup  func(*MockUseCase)
		wantStatus int
	}{
		{
			name:      "success",
			paramCode: "ERR_001",
			mockSetup: func(m *MockUseCase) {
				m.On("GetByCode", mock.Anything, "ERR_001").
					Return(&domain.ErrorCode{Code: "ERR_001", Message: "Something failed"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "bad_request_empty_code",
			paramCode: "",
			mockSetup: func(m *MockUseCase) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "internal_error",
			paramCode: "ERR_001",
			mockSetup: func(m *MockUseCase) {
				m.On("GetByCode", mock.Anything, "ERR_001").
					Return(nil, errors.New("not found"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			c.Params = gin.Params{{Key: "code", Value: tt.paramCode}}

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := errorcode.New(mockUC, log)
			ctrl.Get(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
