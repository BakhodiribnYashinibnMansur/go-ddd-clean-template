package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/usecase"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSystemErrorMiddleware_Recovery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		shouldPanic    bool
		panicValue     interface{}
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
			// Setup logger
			mockLogger := logger.New("debug")

			// Setup mock use case
			uc := &usecase.UseCase{}

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
			errorMessage:  "bad request",
		},
		{
			name:          "skip_404_error",
			statusCode:    http.StatusNotFound,
			shouldPersist: false,
			errorMessage:  "not found",
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
			// Setup logger
			mockLogger := logger.New("debug")

			// Setup mock use case
			uc := &usecase.UseCase{}

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
