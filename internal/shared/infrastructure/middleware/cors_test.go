package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/shared/domain/consts"

	"github.com/gin-gonic/gin"
)

func TestCORSMiddleware_ExactOriginMatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	cfg := config.CORS{
		AllowedOrigins:   []string{"https://example.com"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	r.Use(CORSMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if got := w.Header().Get(consts.HeaderAccessControlAllowOrigin); got != "https://example.com" {
		t.Errorf("expected origin https://example.com, got %s", got)
	}
	if got := w.Header().Get(consts.HeaderAccessControlAllowCredentials); got != "true" {
		t.Errorf("expected credentials true, got %s", got)
	}
}

func TestCORSMiddleware_WildcardWithCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	cfg := config.CORS{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	}

	r.Use(CORSMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://app.example.com")
	r.ServeHTTP(w, req)

	// With credentials and wildcard, should reflect the request origin
	if got := w.Header().Get(consts.HeaderAccessControlAllowOrigin); got != "https://app.example.com" {
		t.Errorf("expected reflected origin, got %s", got)
	}
}

func TestCORSMiddleware_WildcardWithoutCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	cfg := config.CORS{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: false,
	}

	r.Use(CORSMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://any.com")
	r.ServeHTTP(w, req)

	if got := w.Header().Get(consts.HeaderAccessControlAllowOrigin); got != "*" {
		t.Errorf("expected *, got %s", got)
	}
}

func TestCORSMiddleware_PreflightRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	cfg := config.CORS{
		AllowedOrigins: []string{"https://example.com"},
		AllowedMethods: []string{"GET", "POST", "PUT"},
		MaxAge:         3600,
	}

	r.Use(CORSMiddleware(cfg))
	r.OPTIONS("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", w.Code)
	}
}

func TestCORSMiddleware_NoMatchingOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	cfg := config.CORS{
		AllowedOrigins: []string{"https://allowed.com"},
	}

	r.Use(CORSMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://notallowed.com")
	r.ServeHTTP(w, req)

	if got := w.Header().Get(consts.HeaderAccessControlAllowOrigin); got != "" {
		t.Errorf("expected empty allow-origin for non-matching origin, got %s", got)
	}
}

func TestCORSMiddleware_ExposedHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	cfg := config.CORS{
		AllowedOrigins: []string{"*"},
		ExposedHeaders: []string{"X-Custom-Header", "X-Another"},
	}

	r.Use(CORSMiddleware(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get(consts.HeaderAccessControlExposeHeaders); got != "X-Custom-Header, X-Another" {
		t.Errorf("expected exposed headers, got %s", got)
	}
}
