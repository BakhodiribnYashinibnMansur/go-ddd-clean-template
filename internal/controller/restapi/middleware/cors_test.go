package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                 string
		corsConfig           config.CORS
		requestOrigin        string
		requestMethod        string
		expectedAllowOrigin  string
		expectedAllowMethods string
		expectedStatus       int
	}{
		{
			name: "wildcard_origin_without_credentials",
			corsConfig: config.CORS{
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET", "POST"},
				AllowedHeaders:   []string{"Content-Type"},
				AllowCredentials: false,
				MaxAge:           3600,
			},
			requestOrigin:        "https://example.com",
			requestMethod:        "GET",
			expectedAllowOrigin:  "*",
			expectedAllowMethods: "GET, POST",
			expectedStatus:       http.StatusOK,
		},
		{
			name: "wildcard_origin_with_credentials",
			corsConfig: config.CORS{
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET", "POST"},
				AllowedHeaders:   []string{"Content-Type"},
				AllowCredentials: true,
				MaxAge:           3600,
			},
			requestOrigin:        "https://example.com",
			requestMethod:        "GET",
			expectedAllowOrigin:  "https://example.com",
			expectedAllowMethods: "GET, POST",
			expectedStatus:       http.StatusOK,
		},
		{
			name: "specific_origin_match",
			corsConfig: config.CORS{
				AllowedOrigins:   []string{"https://example.com", "https://test.com"},
				AllowedMethods:   []string{"GET", "POST", "PUT"},
				AllowedHeaders:   []string{"Content-Type", "Authorization"},
				AllowCredentials: true,
				MaxAge:           7200,
			},
			requestOrigin:        "https://example.com",
			requestMethod:        "POST",
			expectedAllowOrigin:  "https://example.com",
			expectedAllowMethods: "GET, POST, PUT",
			expectedStatus:       http.StatusOK,
		},
		{
			name: "origin_not_in_allowed_list",
			corsConfig: config.CORS{
				AllowedOrigins:   []string{"https://allowed.com"},
				AllowedMethods:   []string{"GET"},
				AllowedHeaders:   []string{"Content-Type"},
				AllowCredentials: false,
				MaxAge:           3600,
			},
			requestOrigin:        "https://notallowed.com",
			requestMethod:        "GET",
			expectedAllowOrigin:  "",
			expectedAllowMethods: "GET",
			expectedStatus:       http.StatusOK,
		},
		{
			name: "preflight_options_request",
			corsConfig: config.CORS{
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET", "POST", "DELETE"},
				AllowedHeaders:   []string{"Content-Type"},
				AllowCredentials: false,
				MaxAge:           1800,
			},
			requestOrigin:        "https://example.com",
			requestMethod:        "OPTIONS",
			expectedAllowOrigin:  "*",
			expectedAllowMethods: "GET, POST, DELETE",
			expectedStatus:       http.StatusNoContent,
		},
		{
			name: "empty_origin_with_wildcard",
			corsConfig: config.CORS{
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET"},
				AllowedHeaders:   []string{"Content-Type"},
				AllowCredentials: true,
				MaxAge:           3600,
			},
			requestOrigin:        "",
			requestMethod:        "GET",
			expectedAllowOrigin:  "*",
			expectedAllowMethods: "GET",
			expectedStatus:       http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup router with CORS middleware
			r := gin.New()
			r.Use(CORSMiddleware(tt.corsConfig))
			r.GET("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "ok")
			})
			r.POST("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "ok")
			})
			r.OPTIONS("/test", func(c *gin.Context) {
				// OPTIONS is handled by CORS middleware
			})

			// Create request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.requestMethod, "/test", nil)
			if tt.requestOrigin != "" {
				req.Header.Set("Origin", tt.requestOrigin)
			}

			// Execute request
			r.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")

			if tt.expectedAllowOrigin != "" {
				assert.Equal(t, tt.expectedAllowOrigin, w.Header().Get("Access-Control-Allow-Origin"), "Allow-Origin mismatch")
			}

			if tt.expectedAllowMethods != "" {
				assert.Equal(t, tt.expectedAllowMethods, w.Header().Get("Access-Control-Allow-Methods"), "Allow-Methods mismatch")
			}

			// Check credentials header if enabled
			if tt.corsConfig.AllowCredentials {
				assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"), "Allow-Credentials should be true")
			}
		})
	}
}
