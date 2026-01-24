package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		method        string
		path          string
		statusCode    int
		responseBody  string
		shouldHaveLog bool
	}{
		{
			name:          "log_successful_get_request",
			method:        "GET",
			path:          "/api/v1/users",
			statusCode:    http.StatusOK,
			responseBody:  "success",
			shouldHaveLog: true,
		},
		{
			name:          "log_successful_post_request",
			method:        "POST",
			path:          "/api/v1/users",
			statusCode:    http.StatusCreated,
			responseBody:  "created",
			shouldHaveLog: true,
		},
		{
			name:          "log_client_error_request",
			method:        "GET",
			path:          "/api/v1/users/invalid",
			statusCode:    http.StatusBadRequest,
			responseBody:  "bad request",
			shouldHaveLog: true,
		},
		{
			name:          "log_server_error_request",
			method:        "POST",
			path:          "/api/v1/users",
			statusCode:    http.StatusInternalServerError,
			responseBody:  "internal error",
			shouldHaveLog: true,
		},
		{
			name:          "log_not_found_request",
			method:        "GET",
			path:          "/api/v1/notfound",
			statusCode:    http.StatusNotFound,
			responseBody:  "not found",
			shouldHaveLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup logger
			mockLogger := logger.New("debug")

			// Setup router
			r := gin.New()
			r.Use(Logger(mockLogger))

			r.Handle(tt.method, tt.path, func(c *gin.Context) {
				// Simulate some processing time
				time.Sleep(10 * time.Millisecond)
				c.String(tt.statusCode, tt.responseBody)
			})

			// Create request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("User-Agent", "test-agent")

			// Execute request
			r.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.statusCode, w.Code)
			assert.Contains(t, w.Body.String(), tt.responseBody)
		})
	}
}

func TestLogger_WithQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		path         string
		queryParams  string
		expectedCode int
	}{
		{
			name:         "log_request_with_query_params",
			path:         "/api/v1/users",
			queryParams:  "page=1&limit=10&sort=name",
			expectedCode: http.StatusOK,
		},
		{
			name:         "log_request_with_special_chars",
			path:         "/api/v1/search",
			queryParams:  "q=test%20query&filter=active",
			expectedCode: http.StatusOK,
		},
		{
			name:         "log_request_without_query_params",
			path:         "/api/v1/users",
			queryParams:  "",
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup logger
			mockLogger := logger.New("debug")

			// Setup router
			r := gin.New()
			r.Use(Logger(mockLogger))

			r.GET(tt.path, func(c *gin.Context) {
				c.String(http.StatusOK, "ok")
			})

			// Create request
			w := httptest.NewRecorder()
			fullPath := tt.path
			if tt.queryParams != "" {
				fullPath += "?" + tt.queryParams
			}
			req, _ := http.NewRequest("GET", fullPath, nil)

			// Execute request
			r.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}
