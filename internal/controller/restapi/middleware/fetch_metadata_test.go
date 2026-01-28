package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestFetchMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		env            string
		enabled        bool
		headers        map[string]string
		method         string
		expectedStatus int
	}{
		{
			name:           "allow_all_in_dev",
			env:            "development",
			enabled:        true,
			headers:        map[string]string{},
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "block_postman_in_prod (missing headers)",
			env:            "production",
			enabled:        true,
			headers:        map[string]string{},
			method:         "GET",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:    "allow_same_origin_in_prod",
			env:     "production",
			enabled: true,
			headers: map[string]string{
				"Sec-Fetch-Site": "same-origin",
			},
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:    "allow_same_site_in_prod",
			env:     "production",
			enabled: true,
			headers: map[string]string{
				"Sec-Fetch-Site": "same-site",
			},
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:    "allow_top_level_navigation_in_prod",
			env:     "production",
			enabled: true,
			headers: map[string]string{
				"Sec-Fetch-Site": "cross-site",
				"Sec-Fetch-Mode": "navigate",
				"Sec-Fetch-Dest": "document",
			},
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:    "block_cross_site_api_in_prod",
			env:     "production",
			enabled: true,
			headers: map[string]string{
				"Sec-Fetch-Site": "cross-site",
				"Sec-Fetch-Mode": "cors",
				"Sec-Fetch-Dest": "empty",
			},
			method:         "POST",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "allow_if_disabled_in_config",
			env:            "production",
			enabled:        false,
			headers:        map[string]string{},
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.App.Environment = tt.env
			cfg.Security.FetchMetadata = tt.enabled

			r := gin.New()
			r.Use(FetchMetadata(cfg))
			r.GET("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "ok")
			})
			r.POST("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "ok")
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, "/test", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
