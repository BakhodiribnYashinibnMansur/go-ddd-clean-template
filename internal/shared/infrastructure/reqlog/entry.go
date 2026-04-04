// Package reqlog captures every incoming HTTP request + response handled by
// the Gin engine and persists them to the http_request_logs table for audit,
// debugging, and support purposes.
//
// The pipeline is: middleware → Sink → Redis buffer → Flusher → PostgreSQL
// COPY FROM. Writes to Redis are fire-and-forget so request latency is never
// blocked by the logging path.
package reqlog

import "time"

// Entry is a single incoming HTTP request/response record.
type Entry struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Query  string `json:"query,omitempty"`
	Route  string `json:"route,omitempty"`

	RequestHeaders  string `json:"request_headers,omitempty"`
	RequestBody     string `json:"request_body,omitempty"`
	RequestBodySize int    `json:"request_body_size"`

	ResponseStatus   int    `json:"response_status"`
	ResponseHeaders  string `json:"response_headers,omitempty"`
	ResponseBody     string `json:"response_body,omitempty"`
	ResponseBodySize int    `json:"response_body_size"`

	DurationMs int    `json:"duration_ms"`
	ClientIP   string `json:"client_ip,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`

	RequestID string `json:"request_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`

	Timestamp time.Time `json:"ts"`
}

// Sink receives Entries for persistence. Implementations MUST be non-blocking
// and fail silently.
type Sink interface {
	Push(e Entry)
}

// NoopSink discards all entries. Used when persistence is disabled.
type NoopSink struct{}

func (NoopSink) Push(Entry) {}
