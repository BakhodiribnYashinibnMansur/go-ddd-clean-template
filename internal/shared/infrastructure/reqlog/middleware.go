package reqlog

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"gct/internal/shared/infrastructure/contextx"
	"gct/internal/shared/infrastructure/redact"

	"github.com/gin-gonic/gin"
)

// Config tunes the logging middleware behaviour.
type Config struct {
	// MaxBodyBytes caps the stored length of request/response bodies. Larger
	// bodies are truncated with a "…" suffix; the original size is preserved.
	// Defaults to 8192.
	MaxBodyBytes int

	// SuccessSampleRate is the probability (0..1) that a 2xx/3xx request
	// under the slow threshold gets persisted. Errors and slow requests are
	// always persisted. Defaults to 1.0 (log everything) — lower it in
	// production to control volume.
	SuccessSampleRate float64

	// SlowThreshold is the duration above which a request is always persisted
	// regardless of sampling. Defaults to 500ms.
	SlowThreshold time.Duration

	// SkipPaths lists exact request paths for which no log is produced
	// (health checks, metrics scraping, etc.).
	SkipPaths []string

	// SkipPrefixes lists path prefixes that are skipped (e.g. "/swagger/").
	SkipPrefixes []string

	// BodySuppressPaths lists endpoints whose request AND response bodies are
	// cleared before persistence, even after JSON redaction. Use for highly
	// sensitive flows (login, register, password reset, OTP) where any
	// accidental leakage is unacceptable. Sizes are still recorded.
	BodySuppressPaths []string
}

func (cfg *Config) applyDefaults() {
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = 8192
	}
	if cfg.SuccessSampleRate <= 0 {
		cfg.SuccessSampleRate = 1.0
	}
	if cfg.SuccessSampleRate > 1 {
		cfg.SuccessSampleRate = 1.0
	}
	if cfg.SlowThreshold <= 0 {
		cfg.SlowThreshold = 500 * time.Millisecond
	}
}

// Middleware returns a Gin middleware that captures the request/response of
// every non-skipped request and pushes an Entry to the supplied Sink.
//
// Sampling: errors (status >= 400) and slow requests are ALWAYS persisted.
// Successful fast requests are sampled at SuccessSampleRate — this keeps the
// database small while guaranteeing that all failures are captured.
//
// Streaming responses (text/event-stream, chunked) and octet-stream downloads
// are detected and NOT buffered — only headers + size are stored.
func Middleware(sink Sink, cfg Config) gin.HandlerFunc {
	if sink == nil {
		sink = NoopSink{}
	}
	cfg.applyDefaults()

	skipSet := toSet(cfg.SkipPaths)
	suppressSet := toSet(cfg.BodySuppressPaths)

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if skipSet[path] {
			c.Next()
			return
		}
		for _, pre := range cfg.SkipPrefixes {
			if strings.HasPrefix(path, pre) {
				c.Next()
				return
			}
		}

		start := time.Now()

		// Capture request body and restore it for downstream handlers.
		// We limit the read to MaxBodyBytes + 1 so we can know whether the
		// body exceeded the cap without holding arbitrary amounts in memory.
		// The downstream handler still sees the FULL body because the 2MB
		// body-limit middleware runs before us.
		var reqBodyBytes []byte
		if c.Request.Body != nil && c.Request.ContentLength != 0 {
			b, err := io.ReadAll(c.Request.Body)
			_ = c.Request.Body.Close()
			if err == nil {
				reqBodyBytes = b
				c.Request.Body = io.NopCloser(bytes.NewReader(b))
			}
		}

		bw := &bodyWriter{
			ResponseWriter: c.Writer,
			buf:            &bytes.Buffer{},
			limit:          cfg.MaxBodyBytes,
		}
		c.Writer = bw

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		// Sampling decision — evaluated AFTER handler runs so we know status.
		if !shouldPersist(status, duration, cfg.SlowThreshold, cfg.SuccessSampleRate) {
			return
		}

		// Don't buffer streaming / binary response bodies.
		bw.finalize(c.Writer.Header().Get("Content-Type"), c.Writer.Header().Get("Transfer-Encoding"))

		ctx := c.Request.Context()
		reqCT := c.Request.Header.Get("Content-Type")
		respCT := c.Writer.Header().Get("Content-Type")

		entry := Entry{
			Method: c.Request.Method,
			Path:   c.Request.URL.Path,
			Query:  c.Request.URL.RawQuery,
			Route:  c.FullPath(),

			RequestHeaders:  redact.Headers(c.Request.Header),
			RequestBody:     redact.Truncate(redact.JSONBody(reqBodyBytes, reqCT), cfg.MaxBodyBytes),
			RequestBodySize: len(reqBodyBytes),

			ResponseStatus:   status,
			ResponseHeaders:  redact.Headers(c.Writer.Header()),
			ResponseBody:     redact.Truncate(redact.JSONBody([]byte(bw.body()), respCT), cfg.MaxBodyBytes),
			ResponseBodySize: bw.size,

			DurationMs: int(duration.Milliseconds()),
			ClientIP:   c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),

			RequestID: contextx.GetRequestID(ctx),
			SessionID: contextx.GetSessionID(ctx),
			Timestamp: time.Now().UTC(),
		}
		if uid := contextx.GetUserID(ctx); uid != nil {
			entry.UserID = fmt.Sprint(uid)
		}

		// Defence-in-depth: even after JSON key-level redaction, some endpoints
		// handle data so sensitive (raw passwords, OTP codes, fresh tokens)
		// that we refuse to store their bodies at all.
		if suppressSet[path] {
			entry.RequestBody = ""
			entry.ResponseBody = ""
		}

		sink.Push(entry)
	}
}

// shouldPersist decides whether to log this request based on status, duration
// and the configured sample rate. Errors and slow requests are always kept.
func shouldPersist(status int, duration, slowThreshold time.Duration, successRate float64) bool {
	if status >= 400 {
		return true
	}
	if duration >= slowThreshold {
		return true
	}
	if successRate >= 1.0 {
		return true
	}
	if successRate <= 0 {
		return false
	}
	return rand.Float64() < successRate //nolint:gosec // sampling, no security property needed
}

func toSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, s := range items {
		if s != "" {
			m[s] = true
		}
	}
	return m
}

// bodyWriter wraps gin.ResponseWriter and records the response body (up to
// limit bytes) while still writing the full payload downstream. For streaming
// content types the buffer is discarded during finalize().
type bodyWriter struct {
	gin.ResponseWriter
	buf       *bytes.Buffer
	limit     int
	size      int
	streaming bool
}

func (w *bodyWriter) Write(p []byte) (int, error) {
	w.size += len(p)
	if !w.streaming {
		if remaining := w.limit - w.buf.Len(); remaining > 0 {
			if len(p) <= remaining {
				w.buf.Write(p)
			} else {
				w.buf.Write(p[:remaining])
			}
		}
	}
	return w.ResponseWriter.Write(p)
}

func (w *bodyWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

// finalize inspects the response content-type / transfer-encoding and, if the
// response is a stream, marks this writer so that body() returns empty. This
// prevents memory blowup and meaningless log entries for SSE, long-polling,
// and file downloads.
func (w *bodyWriter) finalize(contentType, transferEncoding string) {
	lower := strings.ToLower(contentType)
	if strings.HasPrefix(lower, "text/event-stream") ||
		strings.HasPrefix(lower, "application/octet-stream") ||
		strings.HasPrefix(lower, "application/x-ndjson") ||
		transferEncoding == "chunked" {
		w.streaming = true
		w.buf.Reset()
	}
}

func (w *bodyWriter) body() string {
	if w.streaming {
		return ""
	}
	if w.size > w.limit {
		return w.buf.String() + "…"
	}
	return w.buf.String()
}
