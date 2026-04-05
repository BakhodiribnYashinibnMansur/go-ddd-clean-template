package httpclient

import "time"

// Outcome values categorise why an Entry was emitted. They are derivable from
// response_status + duration_ms at query time, so they are NOT persisted as a
// separate column — kept in-memory for metrics/tests.
const (
	OutcomeError   = "error"   // transport failure or status >= 400
	OutcomeSlow    = "slow"    // 2xx/3xx, duration > SlowThreshold
	OutcomeSampled = "sampled" // 2xx/3xx under threshold, picked by sampling
)

// Entry represents a single 3rd-party HTTP call record to be persisted when
// the call fails, runs slow, or is randomly sampled. Stored in external_api_logs.
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

	// Outcome classifies why the entry was emitted. Not persisted to DB —
	// consumers should derive it from status+duration at query time. Populated
	// on the emit path for metrics, logs, and tests.
	Outcome string `json:"outcome,omitempty"`
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
