package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/kernel/consts"

	"github.com/gin-gonic/gin"
)

func TestLoggerMiddleware_SetsRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &mockLog{}
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Logger(l))
	r.GET("/test", func(c *gin.Context) {
		// Verify request ID is set in context
		reqID := c.GetString(consts.CtxKeyRequestID)
		if reqID == "" {
			t.Error("expected request ID to be set in context")
		}
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Verify the X-Request-ID response header is set
	if got := w.Header().Get(consts.HeaderXRequestID); got == "" {
		t.Error("expected X-Request-ID response header to be set")
	}
}

func TestLoggerMiddleware_PreservesExistingRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &mockLog{}
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Logger(l))
	r.GET("/test", func(c *gin.Context) {
		reqID := c.GetString(consts.CtxKeyRequestID)
		if reqID != "existing-id-123" {
			t.Errorf("expected existing request ID, got %s", reqID)
		}
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set(consts.HeaderXRequestID, "existing-id-123")
	r.ServeHTTP(w, req)

	if got := w.Header().Get(consts.HeaderXRequestID); got != "existing-id-123" {
		t.Errorf("expected existing-id-123, got %s", got)
	}
}

func TestLoggerMiddleware_PassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &mockLog{}
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Logger(l))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"created": true})
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}

	// Allow async logging goroutine to run
	time.Sleep(10 * time.Millisecond)
}

func TestGetStatusColor(t *testing.T) {
	tests := []struct {
		status int
		wantFg string
	}{
		{200, "\033[30m"},
		{301, "\033[30m"},
		{400, "\033[30m"},
		{500, "\033[97m"},
		{100, "\033[97m"},
	}

	for _, tt := range tests {
		fg, _ := getStatusColor(tt.status)
		if fg != tt.wantFg {
			t.Errorf("getStatusColor(%d) fg = %q, want %q", tt.status, fg, tt.wantFg)
		}
	}
}

func TestGetMethodStyle(t *testing.T) {
	tests := []struct {
		method    string
		wantLabel string
	}{
		{"GET", " GET "},
		{"POST", " POST "},
		{"PUT", " PUT "},
		{"DELETE", " DEL "},
		{"PATCH", " PATCH "},
		{"HEAD", " HEAD "},
		{"OPTIONS", " OPT "},
		{"UNKNOWN", " ??? "},
	}

	for _, tt := range tests {
		_, _, label := getMethodStyle(tt.method)
		if label != tt.wantLabel {
			t.Errorf("getMethodStyle(%q) label = %q, want %q", tt.method, label, tt.wantLabel)
		}
	}
}
