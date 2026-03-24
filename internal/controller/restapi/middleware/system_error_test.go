package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/domain"
	"gct/internal/usecase"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSystemErrorUC struct {
	mock.Mock
}

func (m *MockSystemErrorUC) Create(ctx context.Context, in *domain.SystemError) error {
	args := m.Called(ctx, in)
	return args.Error(0)
}

func (m *MockSystemErrorUC) Gets(ctx context.Context, in *domain.SystemErrorsFilter) ([]*domain.SystemError, int, error) {
	args := m.Called(ctx, in)
	return args.Get(0).([]*domain.SystemError), args.Int(1), args.Error(2)
}

func (m *MockSystemErrorUC) Resolve(ctx context.Context, id string, resolvedBy *string) error {
	args := m.Called(ctx, id, resolvedBy)
	return args.Error(0)
}

func TestSystemErrorMiddleware_Recovery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		shouldPanic    bool
		panicValue     any
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "no_panic_normal_flow",
			shouldPanic:    false,
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
		{
			name:           "panic_with_string",
			shouldPanic:    true,
			panicValue:     "something went wrong",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "panic_with_error",
			shouldPanic:    true,
			panicValue:     assert.AnError,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "panic_with_nil",
			shouldPanic:    true,
			panicValue:     nil,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock logger
			mockLogger := logger.New("debug")

			// Setup mock use case
			mockSysErrUC := new(MockSystemErrorUC)
			// Only expect Create call if panic occurs
			if tt.shouldPanic {
				mockSysErrUC.On("Create", mock.Anything, mock.Anything).Return(nil)
			}

			uc := &usecase.UseCase{
				Audit: &TestAuditUseCase{
					SystemErrorUC: mockSysErrUC,
				},
			}

			// Create middleware
			sysErrM := NewSystemErrorMiddleware(uc, mockLogger)

			// Setup router
			r := gin.New()
			r.Use(sysErrM.Recovery())

			r.GET("/test", func(c *gin.Context) {
				if tt.shouldPanic {
					panic(tt.panicValue)
				}
				c.String(http.StatusOK, tt.expectedBody)
			})

			// Create request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)

			// Execute request
			r.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.shouldPanic {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestSystemErrorMiddleware_Persist5xx(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		statusCode    int
		shouldPersist bool
		errorMessage  string
	}{
		{
			name:          "persist_500_error",
			statusCode:    http.StatusInternalServerError,
			shouldPersist: true,
			errorMessage:  "internal server error",
		},
		{
			name:          "persist_502_error",
			statusCode:    http.StatusBadGateway,
			shouldPersist: true,
			errorMessage:  "bad gateway",
		},
		{
			name:          "persist_503_error",
			statusCode:    http.StatusServiceUnavailable,
			shouldPersist: true,
			errorMessage:  "service unavailable",
		},
		{
			name:          "skip_400_error",
			statusCode:    http.StatusBadRequest,
			shouldPersist: false,
			errorMessage:  "",
		},
		{
			name:          "skip_404_error",
			statusCode:    http.StatusNotFound,
			shouldPersist: false,
			errorMessage:  "",
		},
		{
			name:          "skip_200_success",
			statusCode:    http.StatusOK,
			shouldPersist: false,
			errorMessage:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock logger
			mockLogger := logger.New("debug")

			// Setup mock use case
			mockSysErrUC := new(MockSystemErrorUC)
			if tt.shouldPersist {
				mockSysErrUC.On("Create", mock.Anything, mock.Anything).Return(nil)
			}

			uc := &usecase.UseCase{
				Audit: &TestAuditUseCase{
					SystemErrorUC: mockSysErrUC,
				},
			}

			// Create middleware
			sysErrM := NewSystemErrorMiddleware(uc, mockLogger)

			// Setup router
			r := gin.New()
			r.Use(sysErrM.Persist5xx())

			r.GET("/test", func(c *gin.Context) {
				if tt.errorMessage != "" {
					c.Error(assert.AnError)
				}
				c.Status(tt.statusCode)
			})

			// Create request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)

			// Execute request
			r.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}
