package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// debugMockLog captures Debugc calls for assertion.
type debugMockLog struct {
	debugcCalls []debugcCall
}

type debugcCall struct {
	msg string
	kv  []any
}

func (m *debugMockLog) Debug(args ...any)                                     {}
func (m *debugMockLog) Debugf(template string, args ...any)                   {}
func (m *debugMockLog) Debugw(msg string, keysAndValues ...any)               {}
func (m *debugMockLog) Info(args ...any)                                      {}
func (m *debugMockLog) Infof(template string, args ...any)                    {}
func (m *debugMockLog) Infow(msg string, keysAndValues ...any)                {}
func (m *debugMockLog) Warn(args ...any)                                      {}
func (m *debugMockLog) Warnf(template string, args ...any)                    {}
func (m *debugMockLog) Warnw(msg string, keysAndValues ...any)                {}
func (m *debugMockLog) Error(args ...any)                                     {}
func (m *debugMockLog) Errorf(template string, args ...any)                   {}
func (m *debugMockLog) Errorw(msg string, keysAndValues ...any)               {}
func (m *debugMockLog) Fatal(args ...any)                                     {}
func (m *debugMockLog) Fatalf(template string, args ...any)                   {}
func (m *debugMockLog) Fatalw(msg string, keysAndValues ...any)               {}
func (m *debugMockLog) Infoc(ctx context.Context, msg string, kv ...any)      {}
func (m *debugMockLog) Warnc(ctx context.Context, msg string, kv ...any)      {}
func (m *debugMockLog) Errorc(ctx context.Context, msg string, kv ...any)     {}
func (m *debugMockLog) Fatalc(ctx context.Context, msg string, kv ...any)     {}

func (m *debugMockLog) Debugc(ctx context.Context, msg string, kv ...any) {
	m.debugcCalls = append(m.debugcCalls, debugcCall{msg: msg, kv: kv})
}

func TestDebugBodyLogsRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	l := &debugMockLog{}
	r := gin.New()
	r.Use(DebugBody(l))
	r.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	body := `{"username":"alice","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if len(l.debugcCalls) == 0 {
		t.Fatal("expected Debugc to be called, got zero calls")
	}

	call := l.debugcCalls[0]
	if call.msg != "request body" {
		t.Errorf("expected msg %q, got %q", "request body", call.msg)
	}

	// Check that the body key-value is present
	found := false
	for i := 0; i+1 < len(call.kv); i += 2 {
		if call.kv[i] == "body" {
			if got, ok := call.kv[i+1].(string); ok && got == body {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("expected logged body to contain %q, kv: %v", body, call.kv)
	}
}

func TestDebugBodySkipsNilBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	l := &debugMockLog{}
	r := gin.New()
	r.Use(DebugBody(l))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if len(l.debugcCalls) != 0 {
		t.Errorf("expected no Debugc calls for nil body GET, got %d", len(l.debugcCalls))
	}
}

func TestDebugBodySkipsZeroContentLength(t *testing.T) {
	gin.SetMode(gin.TestMode)

	l := &debugMockLog{}
	r := gin.New()
	r.Use(DebugBody(l))
	r.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(""))
	req.ContentLength = 0
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if len(l.debugcCalls) != 0 {
		t.Errorf("expected no Debugc calls for zero content length, got %d", len(l.debugcCalls))
	}
}

func TestDebugBodyRestoresBodyForDownstream(t *testing.T) {
	gin.SetMode(gin.TestMode)

	l := &debugMockLog{}
	r := gin.New()
	r.Use(DebugBody(l))

	original := `{"key":"value"}`
	var downstream string
	r.POST("/test", func(c *gin.Context) {
		b, err := io.ReadAll(c.Request.Body)
		if err != nil {
			t.Fatalf("downstream handler failed to read body: %v", err)
		}
		downstream = string(b)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(original))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if downstream != original {
		t.Errorf("downstream body = %q, want %q", downstream, original)
	}
}

func TestDebugBodyTruncatesAtMaxBodyLog(t *testing.T) {
	gin.SetMode(gin.TestMode)

	l := &debugMockLog{}
	r := gin.New()
	r.Use(DebugBody(l))
	r.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Create a body larger than maxBodyLog (4096)
	largeBody := bytes.Repeat([]byte("A"), maxBodyLog+1024)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/octet-stream")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if len(l.debugcCalls) == 0 {
		t.Fatal("expected Debugc to be called for large body")
	}

	call := l.debugcCalls[0]
	for i := 0; i+1 < len(call.kv); i += 2 {
		if call.kv[i] == "body" {
			logged, ok := call.kv[i+1].(string)
			if !ok {
				t.Fatalf("expected body value to be string, got %T", call.kv[i+1])
			}
			if len(logged) > maxBodyLog {
				t.Errorf("logged body length %d exceeds maxBodyLog %d", len(logged), maxBodyLog)
			}
			return
		}
	}
	t.Error("body key not found in logged key-value pairs")
}
