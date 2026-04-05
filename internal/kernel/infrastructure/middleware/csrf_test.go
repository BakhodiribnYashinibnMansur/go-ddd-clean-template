package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/kernel/consts"

	"github.com/gin-gonic/gin"
)

// mockLog implements logger.Log for testing.
type mockLog struct{}

func (m *mockLog) Debug(args ...any)                                     {}
func (m *mockLog) Debugf(template string, args ...any)                   {}
func (m *mockLog) Debugw(msg string, keysAndValues ...any)               {}
func (m *mockLog) Info(args ...any)                                      {}
func (m *mockLog) Infof(template string, args ...any)                    {}
func (m *mockLog) Infow(msg string, keysAndValues ...any)                {}
func (m *mockLog) Warn(args ...any)                                      {}
func (m *mockLog) Warnf(template string, args ...any)                    {}
func (m *mockLog) Warnw(msg string, keysAndValues ...any)                {}
func (m *mockLog) Error(args ...any)                                     {}
func (m *mockLog) Errorf(template string, args ...any)                   {}
func (m *mockLog) Errorw(msg string, keysAndValues ...any)               {}
func (m *mockLog) Fatal(args ...any)                                     {}
func (m *mockLog) Fatalf(template string, args ...any)                   {}
func (m *mockLog) Fatalw(msg string, keysAndValues ...any)               {}
func (m *mockLog) Debugc(ctx context.Context, msg string, kv ...any)     {}
func (m *mockLog) Infoc(ctx context.Context, msg string, kv ...any)      {}
func (m *mockLog) Warnc(ctx context.Context, msg string, kv ...any)      {}
func (m *mockLog) Errorc(ctx context.Context, msg string, kv ...any)     {}
func (m *mockLog) Fatalc(ctx context.Context, msg string, kv ...any)     {}

func TestCSRFMiddleware_SafeMethodBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &mockLog{}

	for _, method := range []string{"GET", "HEAD", "OPTIONS"} {
		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.Use(Middleware(l, "csrf_token"))
		r.Handle(method, "/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(method, "/test", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("[%s] expected 200, got %d", method, w.Code)
		}
	}
}

func TestCSRFMiddleware_MissingCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &mockLog{}
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Middleware(l, "csrf_token"))
	r.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for missing cookie, got %d", w.Code)
	}
}

func TestCSRFMiddleware_MismatchedTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &mockLog{}
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Middleware(l, "csrf_token"))
	r.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "token-abc"})
	req.Header.Set(consts.HeaderXCSRFToken, "token-xyz")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for mismatched tokens, got %d", w.Code)
	}
}

func TestCSRFMiddleware_ValidTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &mockLog{}
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Middleware(l, "csrf_token"))
	r.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "valid-token"})
	req.Header.Set(consts.HeaderXCSRFToken, "valid-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for valid tokens, got %d", w.Code)
	}
}

func TestCSRFMiddleware_EmptyHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &mockLog{}
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Middleware(l, "csrf_token"))
	r.DELETE("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("DELETE", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "token-value"})
	// No X-CSRF-Token header
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for empty header, got %d", w.Code)
	}
}

func TestHybridMiddleware_BearerTokenBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &mockLog{}
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(HybridMiddleware(l, "csrf_token"))
	r.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set("Authorization", "Bearer some-jwt-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for bearer token bypass, got %d", w.Code)
	}
}

func TestHybridMiddleware_NoBearerFallsBackToCSRF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &mockLog{}
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(HybridMiddleware(l, "csrf_token"))
	r.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	// No Authorization header and no CSRF tokens -> should fail
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 when no auth and no CSRF, got %d", w.Code)
	}
}
