// Package redact provides shared helpers for masking sensitive fields in
// HTTP headers and JSON bodies. Used by reqlog (incoming) and httpclient
// (outgoing) to keep logging paths consistent and safe for compliance
// (GDPR / PCI / SOC2).
package redact

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"gct/internal/shared/infrastructure/logger"
)

// RedactedValue is the placeholder written in place of sensitive values.
const RedactedValue = "***"

// sensitiveHeaders (lowercase) are masked wholesale in persisted logs.
var sensitiveHeaders = map[string]bool{
	"authorization":       true,
	"proxy-authorization": true,
	"cookie":              true,
	"set-cookie":          true,
	"x-api-key":           true,
	"x-auth-token":        true,
	"x-csrf-token":        true,
	"x-access-token":      true,
	"x-refresh-token":     true,
}

// Headers renders http.Header as a stable, sorted "Key: value\n..." string
// with sensitive values replaced by RedactedValue. The output is deterministic
// so identical requests produce identical log entries.
func Headers(h http.Header) string {
	if len(h) == 0 {
		return ""
	}
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		var v string
		if sensitiveHeaders[strings.ToLower(k)] {
			v = RedactedValue
		} else {
			v = strings.Join(h.Values(k), ", ")
		}
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(v)
		b.WriteByte('\n')
	}
	return b.String()
}

// JSONBody parses body as JSON (when contentType indicates JSON) and masks
// values whose keys match logger.IsSensitiveKey, recursively through nested
// objects and arrays. Non-JSON or malformed bodies are returned unchanged —
// this function MUST NEVER drop data or panic.
func JSONBody(body []byte, contentType string) string {
	if len(body) == 0 {
		return ""
	}
	if !strings.Contains(strings.ToLower(contentType), "application/json") {
		return string(body)
	}
	var parsed any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return string(body)
	}
	redacted := walk(parsed)
	out, err := json.Marshal(redacted)
	if err != nil {
		return string(body)
	}
	return string(out)
}

func walk(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if logger.IsSensitiveKey(k) {
				t[k] = RedactedValue
				continue
			}
			t[k] = walk(val)
		}
		return t
	case []any:
		for i, item := range t {
			t[i] = walk(item)
		}
		return t
	default:
		return v
	}
}

// Truncate returns s clipped to at most n bytes, appending "…" if truncated.
func Truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
