package httpclient

import "time"

// Entry represents a single 3rd-party HTTP call record to be persisted when
// the call fails (transport error or status >= 400). Stored in external_api_logs.
type Entry struct {
	APIName   string `json:"api_name"`
	Operation string `json:"operation,omitempty"`

	RequestMethod   string `json:"request_method"`
	RequestURL      string `json:"request_url"`
	RequestHeaders  string `json:"request_headers,omitempty"`
	RequestBody     string `json:"request_body,omitempty"`
	RequestBodySize int    `json:"request_body_size"`

	ResponseStatus   int    `json:"response_status"`
	ResponseHeaders  string `json:"response_headers,omitempty"`
	ResponseBody     string `json:"response_body,omitempty"`
	ResponseBodySize int    `json:"response_body_size"`

	ErrorText  string `json:"error_text,omitempty"`
	DurationMs int    `json:"duration_ms"`

	RequestID string `json:"request_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`

	Timestamp time.Time `json:"ts"`
}

// Sink is the destination where Entries are pushed when an external call fails.
// Implementations MUST be non-blocking and fail silently — logging must never
// affect the caller's hot path.
type Sink interface {
	Push(e Entry)
}

// NoopSink discards all entries. Used when Redis/persistence is disabled.
type NoopSink struct{}

func (NoopSink) Push(Entry) {}
