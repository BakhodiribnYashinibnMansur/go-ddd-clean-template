// Package httpclient provides a logging HTTP client wrapper for 3rd-party API
// calls. By default it captures the full request/response (method, URL,
// headers, body, status, duration) and pushes an Entry to a Sink for
// persistence into the external_api_logs table whenever a call FAILS
// (transport error or status >= 400), runs SLOWER than SlowThreshold, or is
// randomly picked by SuccessSampleRate.
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"gct/internal/kernel/infrastructure/contextx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/redact"
)

const (
	defaultTimeout         = 30 * time.Second
	defaultMaxBodyBytes    = 8192
	defaultSlowThresholdMs = 2000
	// defaultMaxRespRead caps how much of the response body we will read into
	// memory. Protects us against hostile / misbehaving 3rd-party servers that
	// return unbounded payloads.
	defaultMaxRespRead = 10 * 1024 * 1024 // 10 MiB
	// defaultMaxReqRead caps how much of the request body we will buffer into
	// memory for logging. Larger request bodies are still SENT but will be
	// truncated in the log entry.
	defaultMaxReqRead = 10 * 1024 * 1024 // 10 MiB
)

// Options configures a Client.
type Options struct {
	// APIName identifies the 3rd-party service (e.g. "telegram", "stripe").
	// Stored in external_api_logs.api_name.
	APIName string
	// Timeout for the underlying *http.Client. Defaults to 30s.
	Timeout time.Duration
	// MaxBodyBytes is the maximum byte length of request/response bodies stored
	// in the log entry. Larger bodies are truncated with a "…" suffix. The
	// original size is preserved in request_body_size / response_body_size.
	// Defaults to 8192.
	MaxBodyBytes int
	// SlowThreshold is the wall-clock duration above which an otherwise
	// successful (2xx/3xx) response is also persisted. Useful for detecting
	// slow upstream services. Defaults to 2s.
	SlowThreshold time.Duration
	// SuccessSampleRate is the probability (0..1) that a successful, fast
	// call will be persisted for baseline/trend analysis. 0 = never sample
	// (default), 1 = always sample.
	SuccessSampleRate float64
}

// Client is a logging-aware HTTP client for 3rd-party API calls.
type Client struct {
	http              *http.Client
	apiName           string
	sink              Sink
	log               logger.Log
	maxBody           int
	slowThreshold     time.Duration
	successSampleRate float64
}

// New constructs a logging HTTP client. If sink is nil, entries are discarded
// (equivalent to NoopSink). If log is nil, error logging is skipped.
func New(opts Options, sink Sink, log logger.Log) *Client {
	if opts.Timeout <= 0 {
		opts.Timeout = defaultTimeout
	}
	if opts.MaxBodyBytes <= 0 {
		opts.MaxBodyBytes = defaultMaxBodyBytes
	}
	if opts.SlowThreshold <= 0 {
		opts.SlowThreshold = defaultSlowThresholdMs * time.Millisecond
	}
	if opts.SuccessSampleRate < 0 {
		opts.SuccessSampleRate = 0
	}
	if opts.SuccessSampleRate > 1 {
		opts.SuccessSampleRate = 1
	}
	if sink == nil {
		sink = NoopSink{}
	}
	return &Client{
		http:              &http.Client{Timeout: opts.Timeout},
		apiName:           opts.APIName,
		sink:              sink,
		log:               log,
		maxBody:           opts.MaxBodyBytes,
		slowThreshold:     opts.SlowThreshold,
		successSampleRate: opts.SuccessSampleRate,
	}
}

// Do executes req and returns the response, its fully-read body, and any
// error. The caller MUST NOT call resp.Body.Close — it has already been
// consumed and closed. An Entry is pushed to the configured Sink when the
// call fails, exceeds SlowThreshold, or wins the success-sampling dice roll.
//
// The op argument is a caller-supplied operation name (e.g. "SendMessage",
// "CreateCharge") persisted alongside the entry for easier grepping.
func (c *Client) Do(ctx context.Context, req *http.Request, op string) (*http.Response, []byte, error) {
	if ctx == nil {
		// Defensive fallback: callers are expected to pass a context, but we
		// must not nil-panic if one forgets. http.Request requires a context.
		ctx = context.Background()
	}
	req = req.WithContext(ctx)

	// Clone the request body so we can log it AND let http.Client consume it.
	var reqBodyBytes []byte
	if req.Body != nil {
		b, err := io.ReadAll(io.LimitReader(req.Body, defaultMaxReqRead))
		_ = req.Body.Close()
		if err != nil {
			return nil, nil, fmt.Errorf("httpclient: read request body: %w", err)
		}
		reqBodyBytes = b
		req.Body = io.NopCloser(bytes.NewReader(b))
		req.ContentLength = int64(len(b))
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(b)), nil
		}
	}

	start := time.Now()
	resp, doErr := c.http.Do(req)
	duration := time.Since(start)

	var (
		respBodyBytes []byte
		respStatus    int
		respHeaders   http.Header
	)
	if resp != nil {
		respStatus = resp.StatusCode
		respHeaders = resp.Header
		if resp.Body != nil {
			b, readErr := io.ReadAll(io.LimitReader(resp.Body, defaultMaxRespRead))
			_ = resp.Body.Close()
			if readErr != nil && doErr == nil {
				doErr = fmt.Errorf("httpclient: read response body: %w", readErr)
			}
			respBodyBytes = b
			resp.Body = io.NopCloser(bytes.NewReader(b))
		}
	}

	if outcome := c.classify(doErr, respStatus, duration); outcome != "" {
		c.emit(ctx, req, reqBodyBytes, resp, respHeaders, respBodyBytes, respStatus, duration, op, doErr, outcome)
	}

	return resp, respBodyBytes, doErr
}

// classify decides whether (and why) a call should be persisted. Returns an
// empty string to skip emission.
//
//	- transport error or status >= 400 → "error" (always emit)
//	- status < 400 and duration > SlowThreshold → "slow" (always emit)
//	- status < 400 and duration <= SlowThreshold → "sampled" with probability
//	   SuccessSampleRate, else skip
func (c *Client) classify(doErr error, status int, duration time.Duration) string {
	if doErr != nil || status >= 400 {
		return OutcomeError
	}
	if duration > c.slowThreshold {
		return OutcomeSlow
	}
	if c.successSampleRate >= 1.0 {
		return OutcomeSampled
	}
	if c.successSampleRate <= 0 {
		return ""
	}
	if rand.Float64() < c.successSampleRate { //nolint:gosec // sampling, no security property needed
		return OutcomeSampled
	}
	return ""
}

// PostJSON marshals payload to JSON, POSTs it to url, and returns the response,
// its fully-read body, and any error. See Do for logging semantics.
func (c *Client) PostJSON(ctx context.Context, url, op string, payload any) (*http.Response, []byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("httpclient: marshal payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("httpclient: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return c.Do(ctx, req, op)
}

// GetJSON sends a GET to url and returns the response, its body, and any error.
// See Do for logging semantics.
func (c *Client) GetJSON(ctx context.Context, url, op string) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("httpclient: new request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	return c.Do(ctx, req, op)
}

// emit constructs an Entry from a call and pushes it to the sink. The
// outcome parameter records WHY this entry was emitted.
func (c *Client) emit(
	ctx context.Context,
	req *http.Request,
	reqBody []byte,
	resp *http.Response,
	respHeaders http.Header,
	respBody []byte,
	status int,
	duration time.Duration,
	op string,
	doErr error,
	outcome string,
) {
	reqCT := req.Header.Get("Content-Type")
	respCT := ""
	if resp != nil {
		respCT = resp.Header.Get("Content-Type")
	}

	entry := Entry{
		APIName:          c.apiName,
		Operation:        op,
		RequestMethod:    req.Method,
		RequestURL:       req.URL.String(),
		RequestHeaders:   redact.Headers(req.Header),
		RequestBody:      redact.Truncate(redact.JSONBody(reqBody, reqCT), c.maxBody),
		RequestBodySize:  len(reqBody),
		ResponseStatus:   status,
		ResponseBody:     redact.Truncate(redact.JSONBody(respBody, respCT), c.maxBody),
		ResponseBodySize: len(respBody),
		DurationMs:       int(duration.Milliseconds()),
		Timestamp:        time.Now().UTC(),
		Outcome:          outcome,
	}
	if resp != nil {
		entry.ResponseHeaders = redact.Headers(respHeaders)
	}
	if doErr != nil {
		entry.ErrorText = doErr.Error()
	}

	entry.RequestID = contextx.GetRequestID(ctx)
	entry.SessionID = contextx.GetSessionID(ctx)
	entry.IPAddress = contextx.GetIPAddress(ctx)
	if uid := contextx.GetUserID(ctx); uid != nil {
		entry.UserID = fmt.Sprint(uid)
	}

	c.sink.Push(entry)
	incEmitted()

	c.logEmitOutcome(ctx, entry, op, outcome, doErr)
}

// logEmitOutcome surfaces the entry to the application logger at the level
// appropriate for the outcome (error/slow/sampled).
func (c *Client) logEmitOutcome(ctx context.Context, entry Entry, op, outcome string, doErr error) {
	if c.log == nil {
		return
	}
	kv := []any{
		"api_name", c.apiName,
		"operation", op,
		"method", entry.RequestMethod,
		"url", entry.RequestURL,
		"response_status", entry.ResponseStatus,
		"duration_ms", entry.DurationMs,
		"outcome", outcome,
	}
	switch outcome {
	case OutcomeError:
		if doErr != nil {
			kv = append(kv, "error", doErr)
		}
		c.log.Errorc(ctx, "external api error", kv...)
	case OutcomeSlow:
		c.log.Warnc(ctx, "external api slow response", kv...)
	case OutcomeSampled:
		c.log.Debugc(ctx, "external api call sampled", kv...)
	}
}
