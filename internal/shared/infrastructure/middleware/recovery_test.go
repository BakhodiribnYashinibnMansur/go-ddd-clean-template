package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// recoveryMockLog implements logger.Log and tracks whether Errorw was called.
type recoveryMockLog struct {
	errorwCalled bool
	lastMsg      string
}

func (m *recoveryMockLog) Debug(args ...any)                                     {}
func (m *recoveryMockLog) Debugf(template string, args ...any)                   {}
func (m *recoveryMockLog) Debugw(msg string, keysAndValues ...any)               {}
func (m *recoveryMockLog) Info(args ...any)                                      {}
func (m *recoveryMockLog) Infof(template string, args ...any)                    {}
func (m *recoveryMockLog) Infow(msg string, keysAndValues ...any)                {}
func (m *recoveryMockLog) Warn(args ...any)                                      {}
func (m *recoveryMockLog) Warnf(template string, args ...any)                    {}
func (m *recoveryMockLog) Warnw(msg string, keysAndValues ...any)                {}
func (m *recoveryMockLog) Error(args ...any)                                     {}
func (m *recoveryMockLog) Errorf(template string, args ...any)                   {}
func (m *recoveryMockLog) Errorw(msg string, keysAndValues ...any)               { m.errorwCalled = true; m.lastMsg = msg }
func (m *recoveryMockLog) Fatal(args ...any)                                     {}
func (m *recoveryMockLog) Fatalf(template string, args ...any)                   {}
func (m *recoveryMockLog) Fatalw(msg string, keysAndValues ...any)               {}
func (m *recoveryMockLog) Debugc(ctx context.Context, msg string, kv ...any)     {}
func (m *recoveryMockLog) Infoc(ctx context.Context, msg string, kv ...any)      {}
func (m *recoveryMockLog) Warnc(ctx context.Context, msg string, kv ...any)      {}
func (m *recoveryMockLog) Errorc(ctx context.Context, msg string, kv ...any)     {}
func (m *recoveryMockLog) Fatalc(ctx context.Context, msg string, kv ...any)     {}

func TestRecovery_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &recoveryMockLog{}
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Recovery(l))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if l.errorwCalled {
		t.Error("expected Errorw not to be called when no panic occurs")
	}
}

func TestRecovery_PanicReturns500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &recoveryMockLog{}
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Recovery(l))
	r.GET("/test", func(c *gin.Context) {
		panic("test panic")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestRecovery_PanicLogsError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &recoveryMockLog{}
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Recovery(l))
	r.GET("/test", func(c *gin.Context) {
		panic("something broke")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if !l.errorwCalled {
		t.Error("expected Errorw to be called on panic")
	}
	if l.lastMsg != "panic recovered" {
		t.Errorf("expected log message 'panic recovered', got %q", l.lastMsg)
	}
}

func TestRecovery_PanicWithNilValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &recoveryMockLog{}
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Recovery(l))
	r.GET("/test", func(c *gin.Context) {
		panic(nil)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	// Gin's CustomRecovery may or may not catch panic(nil) depending on Go version.
	// At minimum, the server should not crash.
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("expected 200 or 500, got %d", w.Code)
	}
}

func TestRecovery_PanicWithIntValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := &recoveryMockLog{}
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(Recovery(l))
	r.GET("/test", func(c *gin.Context) {
		panic(42)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
	if !l.errorwCalled {
		t.Error("expected Errorw to be called on int panic")
	}
}
