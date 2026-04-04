# Error Body Logging Middleware — Design

## Problem

When a request fails (4xx/5xx), we currently log method, path, status, latency, and the error string — but not what the client actually sent. Debugging user-facing errors (e.g., "why did user creation fail?") requires knowing the request payload.

Existing `DebugBody` middleware logs every request body, but only at debug level. That's too noisy for production and unavailable when we need it most (investigating errors in prod logs).

## Goal

When a request's response status is `>= 400`, log the request body alongside the request metadata so operators can see exactly what the client sent.

## Non-Goals

- Logging response bodies (only request)
- Replacing `DebugBody` (keep it for dev use)
- Logging multipart/file uploads (skip non-JSON bodies)

## Design

### New middleware: `ErrorBody`

Location: `internal/shared/infrastructure/middleware/error_body.go`

**Behavior:**
1. On entry: if method has no body (GET/HEAD/DELETE without body) or `Content-Length == 0` → skip entirely, call `c.Next()`.
2. If `Content-Type` is not JSON (`application/json`) → skip body capture. Non-JSON bodies (multipart, form) may contain binary data and are harder to redact safely.
3. Read up to `maxErrorBodyLog = 4096` bytes from body, restore the full original body for downstream handlers (same pattern as `DebugBody`).
4. Call `c.Next()`.
5. After handler runs, check `c.Writer.Status()`:
   - `< 400` → do nothing (body is discarded, no log emitted)
   - `>= 400` → emit a `Warn`-level log with captured body

**Log fields:**
- `method`, `path`, `status`, `client_ip`, `request_id` (standard)
- `body` (string, redacted, truncated marker if it was > 4KB)
- `content_type`

**Redaction:**
The body is JSON. Parse it into `map[string]any`, walk recursively, redact any key matching the existing `sensitiveKeys` set from `logger/redact.go`. If JSON parse fails (malformed payload), log the raw body as-is (it's already an error path — debugging takes precedence).

To avoid duplicating the sensitive-keys list, expose a small helper from the `logger` package: `logger.IsSensitiveKey(key string) bool`. The middleware uses it for recursive JSON redaction.

### Registration

Add to `middleware/setup.go` after `Logger(l)`, always on (no config flag needed — it only emits on errors, so cost is negligible on the happy path):

```go
// 1. Traceability & Logging
handler.Use(Logger(l))

// 1.1 Error body capture (logs request body on 4xx/5xx)
handler.Use(ErrorBody(l))

// 1.2 Debug-level full body logging (dev only)
if cfg.Log.IsDebug() {
    handler.Use(DebugBody(l))
}
```

### Output

Console (colored via existing logger) and file (JSON via existing persist sink) — both are already handled by the shared `logger.Log` instance. No new sinks needed.

Example console line:
```
WARN  request failed with body  method=POST path=/api/v1/users status=400 body={"email":"bad","password":"***"} ...
```

## Testing

Unit test `error_body_test.go`:
1. `2xx` response → no body log emitted
2. `4xx` response with JSON body containing `password` → log emitted, password redacted, other fields intact
3. Non-JSON `Content-Type` → body not captured, but still logs status (without body field)
4. Empty body (`Content-Length: 0`) → middleware short-circuits, no log
5. Malformed JSON → raw body logged (fallback path)
6. Body > 4KB → truncated to 4KB, `truncated=true` field added, downstream handler still receives full body

## Files Touched

- **New:** `internal/shared/infrastructure/middleware/error_body.go`
- **New:** `internal/shared/infrastructure/middleware/error_body_test.go`
- **Edit:** `internal/shared/infrastructure/logger/redact.go` — expose `IsSensitiveKey`
- **Edit:** `internal/shared/infrastructure/middleware/setup.go` — register middleware
