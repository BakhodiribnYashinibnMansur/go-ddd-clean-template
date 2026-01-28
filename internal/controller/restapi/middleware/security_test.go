package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("it should set security headers", func(t *testing.T) {
		r := gin.New()
		r.Use(Security())
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
		assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
		assert.Equal(t, "none", w.Header().Get("X-Permitted-Cross-Domain-Policies"))
		assert.Contains(t, w.Header().Get("Content-Security-Policy"), "default-src 'self'")
	})

	t.Run("it should set HSTS when HTTPS", func(t *testing.T) {
		r := gin.New()
		r.Use(Security())
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		// Test using X-Forwarded-Proto
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Forwarded-Proto", "https")
		r.ServeHTTP(w, req)
		assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "max-age=31536000")
	})

	t.Run("it should not set HSTS when not HTTPS", func(t *testing.T) {
		r := gin.New()
		r.Use(Security())
		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		r.ServeHTTP(w, req)
		assert.Empty(t, w.Header().Get("Strict-Transport-Security"))
	})
}
