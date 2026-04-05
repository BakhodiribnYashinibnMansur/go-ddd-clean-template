package reqlog

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// captureSink collects every entry pushed to it, for assertion.
type captureSink struct {
	mu      sync.Mutex
	entries []Entry
}

func (s *captureSink) Push(e Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, e)
}

func (s *captureSink) last() Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.entries) == 0 {
		return Entry{}
	}
	return s.entries[len(s.entries)-1]
}

func (s *captureSink) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries)
}

func init() { gin.SetMode(gin.TestMode) }

func newEngine(sink Sink, cfg Config) *gin.Engine {
	e := gin.New()
	e.Use(Middleware(sink, cfg))
	return e
}

func do(e *gin.Engine, method, path string, body []byte, headers http.Header) *httptest.ResponseRecorder {
	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	if body != nil {
		req.ContentLength = int64(len(body))
	}
	for k, vs := range headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	return w
}

func TestMiddleware_CapturesRequestAndResponse(t *testing.T) {
	sink := &captureSink{}
	e := newEngine(sink, Config{})
	e.POST("/echo", func(c *gin.Context) {
		b, _ := io.ReadAll(c.Request.Body)
		c.Data(http.StatusOK, "application/json", b)
	})

	body := []byte(`{"hello":"world"}`)
	h := http.Header{"Content-Type": {"application/json"}}
	w := do(e, http.MethodPost, "/echo", body, h)

	if w.Code != 200 {
		t.Fatalf("status=%d", w.Code)
	}
	if sink.count() != 1 {
		t.Fatalf("expected 1 entry, got %d", sink.count())
	}
	entry := sink.last()
	if entry.Method != "POST" || entry.Path != "/echo" {
		t.Errorf("bad entry: %+v", entry)
	}
	if !strings.Contains(entry.RequestBody, "world") {
		t.Errorf("request body not captured: %q", entry.RequestBody)
	}
	if !strings.Contains(entry.ResponseBody, "world") {
		t.Errorf("response body not captured: %q", entry.ResponseBody)
	}
	if entry.RequestBodySize != len(body) {
		t.Errorf("wrong req size: %d", entry.RequestBodySize)
	}
}

func TestMiddleware_RedactsJSONPassword(t *testing.T) {
	sink := &captureSink{}
	e := newEngine(sink, Config{})
	e.POST("/login", func(c *gin.Context) { c.Status(http.StatusOK) })

	body := []byte(`{"email":"a@b","password":"hunter2"}`)
	h := http.Header{"Content-Type": {"application/json"}}
	do(e, http.MethodPost, "/login", body, h)

	if strings.Contains(sink.last().RequestBody, "hunter2") {
		t.Fatalf("password leaked: %s", sink.last().RequestBody)
	}
}

func TestMiddleware_BodySuppressPaths(t *testing.T) {
	sink := &captureSink{}
	e := newEngine(sink, Config{
		BodySuppressPaths: []string{"/api/v1/auth/login"},
	})
	e.POST("/api/v1/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"token": "secret"})
	})

	body := []byte(`{"email":"a","password":"b"}`)
	do(e, http.MethodPost, "/api/v1/auth/login", body,
		http.Header{"Content-Type": {"application/json"}})

	entry := sink.last()
	if entry.RequestBody != "" || entry.ResponseBody != "" {
		t.Fatalf("bodies not suppressed: req=%q resp=%q",
			entry.RequestBody, entry.ResponseBody)
	}
	if entry.RequestBodySize != len(body) {
		t.Errorf("size not tracked: %d", entry.RequestBodySize)
	}
}

func TestMiddleware_SkipPaths(t *testing.T) {
	sink := &captureSink{}
	e := newEngine(sink, Config{SkipPaths: []string{"/health"}})
	e.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	do(e, http.MethodGet, "/health", nil, nil)
	if sink.count() != 0 {
		t.Fatalf("skipped path was logged: %d entries", sink.count())
	}
}

func TestMiddleware_SkipPrefixes(t *testing.T) {
	sink := &captureSink{}
	e := newEngine(sink, Config{SkipPrefixes: []string{"/swagger/"}})
	e.GET("/swagger/index.html", func(c *gin.Context) { c.Status(http.StatusOK) })
	do(e, http.MethodGet, "/swagger/index.html", nil, nil)
	if sink.count() != 0 {
		t.Fatalf("prefix-matched path was logged")
	}
}

func TestMiddleware_TruncatesLargeBody(t *testing.T) {
	sink := &captureSink{}
	e := newEngine(sink, Config{MaxBodyBytes: 16})
	e.POST("/x", func(c *gin.Context) {
		_, _ = io.Copy(io.Discard, c.Request.Body)
		c.Status(http.StatusOK)
	})
	body := bytes.Repeat([]byte("a"), 100)
	do(e, http.MethodPost, "/x", body, http.Header{"Content-Type": {"text/plain"}})
	entry := sink.last()
	if !strings.HasSuffix(entry.RequestBody, "…") {
		t.Fatalf("expected truncation marker, got %q", entry.RequestBody)
	}
	if entry.RequestBodySize != 100 {
		t.Fatalf("size lost: %d", entry.RequestBodySize)
	}
}

func TestMiddleware_RestoresRequestBodyForHandler(t *testing.T) {
	sink := &captureSink{}
	e := newEngine(sink, Config{})
	var seen []byte
	e.POST("/x", func(c *gin.Context) {
		b, _ := io.ReadAll(c.Request.Body)
		seen = b
		c.Status(http.StatusOK)
	})
	body := []byte(`{"k":"v"}`)
	do(e, http.MethodPost, "/x", body, nil)
	if string(seen) != string(body) {
		t.Fatalf("handler saw %q, want %q", seen, body)
	}
}

func TestMiddleware_SamplesSuccessfulRequests(t *testing.T) {
	sink := &captureSink{}
	// 0.0001 sample rate — should drop ~all successful entries.
	e := newEngine(sink, Config{SuccessSampleRate: 0.0001, SlowThreshold: time.Hour})
	e.GET("/ok", func(c *gin.Context) { c.Status(http.StatusOK) })

	for i := 0; i < 200; i++ {
		do(e, http.MethodGet, "/ok", nil, nil)
	}
	if sink.count() > 5 {
		t.Fatalf("sampling not working: %d entries", sink.count())
	}
}

func TestMiddleware_AlwaysLogsErrors(t *testing.T) {
	sink := &captureSink{}
	e := newEngine(sink, Config{SuccessSampleRate: 0.0001, SlowThreshold: time.Hour})
	e.GET("/fail", func(c *gin.Context) { c.Status(http.StatusInternalServerError) })

	for i := 0; i < 10; i++ {
		do(e, http.MethodGet, "/fail", nil, nil)
	}
	if sink.count() != 10 {
		t.Fatalf("error entries sampled away: %d/10", sink.count())
	}
}

func TestMiddleware_AlwaysLogsSlowRequests(t *testing.T) {
	sink := &captureSink{}
	e := newEngine(sink, Config{SuccessSampleRate: 0.0001, SlowThreshold: 10 * time.Millisecond})
	e.GET("/slow", func(c *gin.Context) {
		time.Sleep(25 * time.Millisecond)
		c.Status(http.StatusOK)
	})
	do(e, http.MethodGet, "/slow", nil, nil)
	if sink.count() != 1 {
		t.Fatalf("slow request not logged: %d entries", sink.count())
	}
}

func TestBodyWriter_SkipsSSE(t *testing.T) {
	sink := &captureSink{}
	e := newEngine(sink, Config{})
	e.GET("/sse", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		_, _ = c.Writer.Write([]byte("data: hello\n\n"))
		_, _ = c.Writer.Write([]byte("data: world\n\n"))
	})
	do(e, http.MethodGet, "/sse", nil, nil)
	entry := sink.last()
	if entry.ResponseBody != "" {
		t.Fatalf("SSE body was buffered: %q", entry.ResponseBody)
	}
	if entry.ResponseBodySize == 0 {
		t.Fatalf("size should still be recorded for SSE")
	}
}

func TestMiddleware_ResponseBodyIsJSONRedacted(t *testing.T) {
	sink := &captureSink{}
	e := newEngine(sink, Config{})
	e.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"access_token": "tok-abc",
			"user":         gin.H{"name": "Ali"},
		})
	})
	do(e, http.MethodPost, "/login", []byte(`{}`),
		http.Header{"Content-Type": {"application/json"}})

	entry := sink.last()
	if strings.Contains(entry.ResponseBody, "tok-abc") {
		t.Fatalf("access_token leaked: %s", entry.ResponseBody)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(entry.ResponseBody), &parsed); err != nil {
		t.Fatalf("redacted response not valid JSON: %v (%s)", err, entry.ResponseBody)
	}
}

func TestMiddleware_RedactsAuthorizationHeader(t *testing.T) {
	sink := &captureSink{}
	e := newEngine(sink, Config{})
	e.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })
	do(e, http.MethodGet, "/x", nil, http.Header{"Authorization": {"Bearer leaky"}})
	if strings.Contains(sink.last().RequestHeaders, "leaky") {
		t.Fatalf("authorization leaked: %s", sink.last().RequestHeaders)
	}
}
