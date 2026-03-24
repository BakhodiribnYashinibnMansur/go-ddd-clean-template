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

func TestGets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		mockSetup  func(*MockUseCase)
		wantStatus int
	}{
		{
			name: "success",
			mockSetup: func(m *MockUseCase) {
				m.On("List", mock.Anything).
					Return([]*domain.ErrorCode{{Code: "ERR_001"}}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "success_empty",
			mockSetup: func(m *MockUseCase) {
				m.On("List", mock.Anything).
					Return([]*domain.ErrorCode{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "internal_error",
			mockSetup: func(m *MockUseCase) {
				m.On("List", mock.Anything).
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

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := errorcode.New(mockUC, log)
			ctrl.Gets(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
