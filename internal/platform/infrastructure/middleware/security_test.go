package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gct/internal/platform/domain/consts"

	"github.com/gin-gonic/gin"
)

func TestSecurity_SetsXContentTypeOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Security())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get(consts.HeaderXContentTypeOptions); got != consts.HeaderValueNoSniff {
		t.Errorf("expected %q, got %q", consts.HeaderValueNoSniff, got)
	}
}

func TestSecurity_SetsXFrameOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Security())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get(consts.HeaderXFrameOptions); got != consts.HeaderValueDeny {
		t.Errorf("expected %q, got %q", consts.HeaderValueDeny, got)
	}
}

func TestSecurity_SetsXXSSProtection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Security())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get(consts.HeaderXXSSProtection); got != consts.HeaderValueXSSBlock {
		t.Errorf("expected %q, got %q", consts.HeaderValueXSSBlock, got)
	}
}

func TestSecurity_SetsReferrerPolicy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Security())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get(consts.HeaderReferrerPolicy); got != consts.HeaderValueStrictOrigin {
		t.Errorf("expected %q, got %q", consts.HeaderValueStrictOrigin, got)
	}
}

func TestSecurity_SetsXPermittedCrossDomainPolicies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Security())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get(consts.HeaderXPermittedCP); got != consts.HeaderValueNone {
		t.Errorf("expected %q, got %q", consts.HeaderValueNone, got)
	}
}

func TestSecurity_SetsCSP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Security())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	csp := w.Header().Get(consts.HeaderCSP)
	if csp == "" {
		t.Fatal("expected Content-Security-Policy header to be set")
	}
	if !strings.Contains(csp, "default-src 'self'") {
		t.Errorf("expected CSP to contain default-src 'self', got %q", csp)
	}
	if !strings.Contains(csp, "script-src") {
		t.Errorf("expected CSP to contain script-src directive, got %q", csp)
	}
	if !strings.Contains(csp, "frame-ancestors 'none'") {
		t.Errorf("expected CSP to contain frame-ancestors 'none', got %q", csp)
	}
}

func TestSecurity_NoHSTSForPlainHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Security())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest("GET", "/test", nil)
	// Plain HTTP, no TLS, no X-Forwarded-Proto
	r.ServeHTTP(w, req)

	if got := w.Header().Get(consts.HeaderHSTS); got != "" {
		t.Errorf("expected no HSTS header for plain HTTP, got %q", got)
	}
}

func TestSecurity_HSTSForXForwardedProtoHTTPS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Security())
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set(consts.HeaderXForwardedProto, "https")
	r.ServeHTTP(w, req)

	if got := w.Header().Get(consts.HeaderHSTS); got != consts.HeaderValueHSTS {
		t.Errorf("expected HSTS %q, got %q", consts.HeaderValueHSTS, got)
	}
}

func TestSecurity_PassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Security())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestSecurityCustom_UsesCustomCSP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	customCSP := []string{"default-src 'none'", "img-src 'self'"}
	r.Use(SecurityCustom(customCSP))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	got := w.Header().Get(consts.HeaderCSP)
	expected := "default-src 'none'; img-src 'self'"
	if got != expected {
		t.Errorf("expected CSP %q, got %q", expected, got)
	}
}

func TestSecurityCustom_EmptyCSP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(SecurityCustom(nil))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if got := w.Header().Get(consts.HeaderCSP); got != "" {
		t.Errorf("expected empty CSP for nil directives, got %q", got)
	}

	// Other security headers should still be present
	if got := w.Header().Get(consts.HeaderXFrameOptions); got != consts.HeaderValueDeny {
		t.Errorf("expected X-Frame-Options %q, got %q", consts.HeaderValueDeny, got)
	}
}

func TestSecurityCustom_HSTSBehavior(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(SecurityCustom([]string{"default-src 'self'"}))
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set(consts.HeaderXForwardedProto, "https")
	r.ServeHTTP(w, req)

	if got := w.Header().Get(consts.HeaderHSTS); got != consts.HeaderValueHSTS {
		t.Errorf("expected HSTS %q, got %q", consts.HeaderValueHSTS, got)
	}
}
