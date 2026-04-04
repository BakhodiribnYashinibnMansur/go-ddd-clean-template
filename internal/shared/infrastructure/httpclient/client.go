// Package httpclient provides a logging HTTP client wrapper for 3rd-party API
// calls. On error (transport failure or response status >= 400) it captures
// the full request/response (method, URL, headers, body, status, duration) and
// pushes an Entry to a Sink for persistence into the external_api_logs table.
//
// Successful calls produce no persisted record, keeping the log table focused
// on debugging failures: "with THIS payload, THIS error occurred".
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gct/internal/shared/infrastructure/contextx"
	"gct/internal/shared/infrastructure/logger"
	"gct/internal/shared/infrastructure/redact"
)

const (
	defaultTimeout      = 30 * time.Second
	defaultMaxBodyBytes = 8192
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
}

// Client is a logging-aware HTTP client for 3rd-party API calls.
type Client struct {
	http    *http.Client
	apiName string
	sink    Sink
	log     logger.Log
	maxBody int
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
	if sink == nil {
		sink = NoopSink{}
	}
	return &Client{
		http:    &http.Client{Timeout: opts.Timeout},
		apiName: opts.APIName,
		sink:    sink,
		log:     log,
		maxBody: opts.MaxBodyBytes,
	}
}

// Do executes req and returns the response, its fully-read body, and any
// error. The caller MUST NOT call resp.Body.Close — it has already been
// consumed and closed. On transport error or status >= 400, an Entry is
// pushed to the configured Sink.
//
// The op argument is a caller-supplied operation name (e.g. "SendMessage",
// "CreateCharge") persisted alongside the entry for easier grepping.
func (c *Client) Do(ctx context.Context, req *http.Request, op string) (*http.Response, []byte, error) {
	if ctx == nil {
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

	isErr := doErr != nil || respStatus >= 400
	if isErr {
		c.emit(ctx, req, reqBodyBytes, resp, respHeaders, respBodyBytes, respStatus, duration, op, doErr)
	}

	return resp, respBodyBytes, doErr
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

// emit constructs an Entry from a failed call and pushes it to the sink.
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

	// Surface the error to the application logger for immediate visibility.
	if c.log != nil {
		kv := []any{
			"api_name", c.apiName,
			"operation", op,
			"method", entry.RequestMethod,
			"url", entry.RequestURL,
			"response_status", status,
			"duration_ms", entry.DurationMs,
		}
		if doErr != nil {
			kv = append(kv, "error", doErr)
		}
		c.log.Errorc(ctx, "external api error", kv...)
	}
}
