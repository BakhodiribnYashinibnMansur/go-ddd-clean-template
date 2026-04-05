package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

// captureLog records Warnc calls so tests can assert on what was logged.
// All other logger methods are no-ops.
type captureLog struct {
	mu     sync.Mutex
	warns  []logEntry
	logins []logEntry
}

type logEntry struct {
	msg    string
	fields map[string]any
}

func (c *captureLog) Warnc(_ context.Context, msg string, kv ...any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.warns = append(c.warns, logEntry{msg: msg, fields: kvToMap(kv)})
}

// last returns the most recent Warnc entry, or a zero entry if none.
func (c *captureLog) last() logEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.warns) == 0 {
		return logEntry{}
	}
	return c.warns[len(c.warns)-1]
}

func (c *captureLog) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.warns)
}

func kvToMap(kv []any) map[string]any {
	m := make(map[string]any, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		key, _ := kv[i].(string)
		m[key] = kv[i+1]
	}
	return m
}

// All remaining Log interface methods — no-ops.
func (c *captureLog) Debug(...any)                           {}
func (c *captureLog) Debugf(string, ...any)                  {}
func (c *captureLog) Debugw(string, ...any)                  {}
func (c *captureLog) Info(...any)                            {}
func (c *captureLog) Infof(string, ...any)                   {}
func (c *captureLog) Infow(string, ...any)                   {}
func (c *captureLog) Warn(...any)                            {}
func (c *captureLog) Warnf(string, ...any)                   {}
func (c *captureLog) Warnw(string, ...any)                   {}
func (c *captureLog) Error(...any)                           {}
func (c *captureLog) Errorf(string, ...any)                  {}
func (c *captureLog) Errorw(string, ...any)                  {}
func (c *captureLog) Fatal(...any)                           {}
func (c *captureLog) Fatalf(string, ...any)                  {}
func (c *captureLog) Fatalw(string, ...any)                  {}
func (c *captureLog) Debugc(context.Context, string, ...any) {}
func (c *captureLog) Infoc(context.Context, string, ...any)  {}
func (c *captureLog) Errorc(context.Context, string, ...any) {}
func (c *captureLog) Fatalc(context.Context, string, ...any) {}

func newRouter(l *captureLog, handler gin.HandlerFunc) (*gin.Engine, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.Use(ErrorBody(l))
	r.POST("/users", handler)
	return r, w
}

func TestErrorBody_SuccessResponseDoesNotLog(t *testing.T) {
	l := &captureLog{}
	r, w := newRouter(l, func(c *gin.Context) {
		// Drain body to emulate a real handler.
		_, _ = io.ReadAll(c.Request.Body)
		c.Status(http.StatusCreated)
	})

	req, _ := http.NewRequest("POST", "/users", strings.NewReader(`{"email":"a@b.co"}`))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(`{"email":"a@b.co"}`))
	r.ServeHTTP(w, req)

	if l.count() != 0 {
		t.Fatalf("expected no logs on 2xx, got %d", l.count())
	}
}

func TestErrorBody_ErrorResponseLogsBodyWithRedaction(t *testing.T) {
	l := &captureLog{}
	r, w := newRouter(l, func(c *gin.Context) {
		_, _ = io.ReadAll(c.Request.Body)
		c.Status(http.StatusBadRequest)
	})

	payload := `{"email":"bad","password":"hunter2","nested":{"token":"abc","name":"x"}}`
	req, _ := http.NewRequest("POST", "/users", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(payload))
	r.ServeHTTP(w, req)

	if l.count() != 1 {
		t.Fatalf("expected 1 log, got %d", l.count())
	}
	entry := l.last()
	bodyStr, _ := entry.fields["body"].(string)
	if bodyStr == "" {
		t.Fatalf("expected body field, got entry=%+v", entry)
	}

	// Decode and verify redaction.
	var parsed map[string]any
	if err := json.Unmarshal([]byte(bodyStr), &parsed); err != nil {
		t.Fatalf("body not valid JSON: %v (body=%q)", err, bodyStr)
	}
	if parsed["password"] != "***" {
		t.Errorf("password not redacted: %v", parsed["password"])
	}
	if parsed["email"] != "bad" {
		t.Errorf("email was altered: %v", parsed["email"])
	}
	nested, _ := parsed["nested"].(map[string]any)
	if nested["token"] != "***" {
		t.Errorf("nested token not redacted: %v", nested["token"])
	}
	if nested["name"] != "x" {
		t.Errorf("nested.name altered: %v", nested["name"])
	}

	if entry.fields["status"] != 400 {
		t.Errorf("expected status=400, got %v", entry.fields["status"])
	}
}

func TestErrorBody_NonJSONBodyNotCaptured(t *testing.T) {
	l := &captureLog{}
	r, w := newRouter(l, func(c *gin.Context) {
		_, _ = io.ReadAll(c.Request.Body)
		c.Status(http.StatusBadRequest)
	})

	req, _ := http.NewRequest("POST", "/users", strings.NewReader("name=foo&password=secret"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ContentLength = 24
	r.ServeHTTP(w, req)

	if l.count() != 1 {
		t.Fatalf("expected 1 log, got %d", l.count())
	}
	entry := l.last()
	if _, ok := entry.fields["body"]; ok {
		t.Errorf("non-JSON body should not be captured, got body=%v", entry.fields["body"])
	}
}

func TestErrorBody_EmptyBodyShortCircuits(t *testing.T) {
	l := &captureLog{}
	r, w := newRouter(l, func(c *gin.Context) {
		c.Status(http.StatusBadRequest)
	})

	req, _ := http.NewRequest("POST", "/users", nil)
	r.ServeHTTP(w, req)

	if l.count() != 1 {
		t.Fatalf("expected 1 log (status only), got %d", l.count())
	}
	if _, ok := l.last().fields["body"]; ok {
		t.Errorf("empty body should not produce body field")
	}
}

func TestErrorBody_MalformedJSONLoggedRaw(t *testing.T) {
	l := &captureLog{}
	r, w := newRouter(l, func(c *gin.Context) {
		_, _ = io.ReadAll(c.Request.Body)
		c.Status(http.StatusBadRequest)
	})

	payload := `{"email": "bad" // not json`
	req, _ := http.NewRequest("POST", "/users", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(payload))
	r.ServeHTTP(w, req)

	entry := l.last()
	if entry.fields["body"] != payload {
		t.Errorf("expected raw body on malformed JSON, got %v", entry.fields["body"])
	}
}

func TestErrorBody_LargeBodyTruncatedButHandlerSeesFull(t *testing.T) {
	l := &captureLog{}
	// Build a JSON body larger than maxErrorBodyLog (4096 bytes).
	big := make([]byte, 0, 6000)
	big = append(big, `{"data":"`...)
	big = append(big, bytes.Repeat([]byte("x"), 5500)...)
	big = append(big, `"}`...)

	var handlerSaw int
	r, w := newRouter(l, func(c *gin.Context) {
		b, _ := io.ReadAll(c.Request.Body)
		handlerSaw = len(b)
		c.Status(http.StatusBadRequest)
	})

	req, _ := http.NewRequest("POST", "/users", bytes.NewReader(big))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(big))
	r.ServeHTTP(w, req)

	if handlerSaw != len(big) {
		t.Errorf("handler got %d bytes, expected full %d", handlerSaw, len(big))
	}
	entry := l.last()
	if entry.fields["truncated"] != true {
		t.Errorf("expected truncated=true, got %v", entry.fields["truncated"])
	}
}
