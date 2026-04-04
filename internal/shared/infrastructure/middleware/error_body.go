package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"gct/internal/shared/domain/consts"
	"gct/internal/shared/infrastructure/httpx"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

const maxErrorBodyLog = 4096 // max bytes captured from request body for error logging

// ErrorBody captures the request body (JSON only, truncated to 4KB) and,
// if the handler ultimately responds with a 4xx/5xx status, emits a
// warn-level log containing the body so operators can see what the client
// sent. Sensitive fields inside the JSON payload (password, token, etc.)
// are redacted using the logger's sensitive-key list.
//
// On 2xx/3xx responses nothing is logged — this middleware is cheap on the
// happy path and only adds observability where it matters.
func ErrorBody(l logger.Log) gin.HandlerFunc {
	return func(c *gin.Context) {
		captured, truncated := captureJSONBody(c)

		c.Next()

		status := c.Writer.Status()
		if status < 400 {
			return
		}

		ctx := c.Request.Context()
		requestID := c.GetString(consts.CtxKeyRequestID)

		fields := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"client_ip", httpx.GetIPAddress(c),
			consts.CtxKeyRequestID, requestID,
			"content_type", c.Request.Header.Get("Content-Type"),
		}
		if captured != "" {
			fields = append(fields, "body", captured)
			if truncated {
				fields = append(fields, "truncated", true)
			}
		}

		l.Warnc(ctx, "request failed with body", fields...)
	}
}

// captureJSONBody reads up to maxErrorBodyLog bytes from a JSON request body,
// restores the original body stream for downstream handlers, and returns the
// (redacted) string form along with a truncated flag. Returns ("", false)
// for non-JSON or empty bodies.
func captureJSONBody(c *gin.Context) (string, bool) {
	if c.Request.Body == nil || c.Request.ContentLength == 0 {
		return "", false
	}
	ct := c.Request.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(ct), "application/json") {
		return "", false
	}

	// Read up to the limit, then drain the remainder so we can rebuild
	// the full body for downstream handlers without losing any bytes.
	head, err := io.ReadAll(io.LimitReader(c.Request.Body, maxErrorBodyLog))
	if err != nil {
		return "", false
	}
	rest, _ := io.ReadAll(c.Request.Body)

	// Restore full body for the handler chain.
	c.Request.Body = io.NopCloser(bytes.NewReader(append(append([]byte{}, head...), rest...)))

	truncated := len(rest) > 0

	// Try to parse as JSON and redact sensitive fields. If parsing fails
	// (malformed JSON), fall back to the raw captured bytes — this is
	// already an error path and visibility trumps strictness.
	var parsed any
	if err := json.Unmarshal(head, &parsed); err != nil {
		return string(head), truncated
	}
	redacted := redactJSON(parsed)
	out, err := json.Marshal(redacted)
	if err != nil {
		return string(head), truncated
	}
	return string(out), truncated
}

// redactJSON walks a decoded JSON value and replaces the value of any
// sensitive key with "***". Works recursively through nested objects
// and arrays.
func redactJSON(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if logger.IsSensitiveKey(k) {
				t[k] = "***"
				continue
			}
			t[k] = redactJSON(val)
		}
		return t
	case []any:
		for i, item := range t {
			t[i] = redactJSON(item)
		}
		return t
	default:
		return v
	}
}
