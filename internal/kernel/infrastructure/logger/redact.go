package logger

import (
	"fmt"
	"strings"

	"go.uber.org/zap/zapcore"
)

// sensitiveKeys are field names whose values should be redacted in logs.
var sensitiveKeys = map[string]bool{
	"password":      true,
	"token":         true,
	"access_token":  true,
	"refresh_token": true,
	"api_key":       true,
	"secret":        true,
	"authorization": true,
	"cookie":        true,
	"csrf_token":    true,
	"otp":           true,
	"pin":           true,
}

const redactedValue = "***"

// redactCore wraps a zapcore.Core and redacts sensitive field values.
type redactCore struct {
	zapcore.Core
}

// NewRedactCore wraps a core to automatically redact sensitive fields.
func NewRedactCore(core zapcore.Core) zapcore.Core {
	return &redactCore{Core: core}
}

func (c *redactCore) With(fields []zapcore.Field) zapcore.Core {
	return &redactCore{Core: c.Core.With(redactFields(fields))}
}

func (c *redactCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Core.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *redactCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	if err := c.Core.Write(ent, redactFields(fields)); err != nil {
		return fmt.Errorf("logger.redactCore.Write: %w", err)
	}
	return nil
}

func redactFields(fields []zapcore.Field) []zapcore.Field {
	out := make([]zapcore.Field, len(fields))
	for i, f := range fields {
		if isSensitive(f.Key) {
			out[i] = zapcore.Field{
				Key:    f.Key,
				Type:   zapcore.StringType,
				String: redactedValue,
			}
		} else {
			out[i] = f
		}
	}
	return out
}

func isSensitive(key string) bool {
	lower := strings.ToLower(key)
	return sensitiveKeys[lower]
}

// IsSensitiveKey reports whether a field name matches the configured
// sensitive-keys set (password, token, secret, etc.). Exposed for callers
// (e.g. middleware) that need to redact values inside structured payloads
// such as JSON request bodies.
func IsSensitiveKey(key string) bool {
	return isSensitive(key)
}
