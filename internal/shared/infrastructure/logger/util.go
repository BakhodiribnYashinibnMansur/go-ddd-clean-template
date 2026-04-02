package logger

import (
	"context"

	"gct/internal/shared/infrastructure/contextx"

	"go.uber.org/zap"
)

// WithFields enriches context with field values.
func WithFields(ctx context.Context, fields map[string]any) context.Context {
	for k, v := range fields {
		switch k {
		case contextx.FieldRequestID:
			if s, ok := v.(string); ok {
				ctx = contextx.WithRequestID(ctx, s)
			}
		case contextx.FieldSessionID:
			if s, ok := v.(string); ok {
				ctx = contextx.WithSessionID(ctx, s)
			}
		case contextx.FieldUserID:
			ctx = contextx.WithUserID(ctx, v)
		case contextx.FieldUserRole:
			if s, ok := v.(string); ok {
				ctx = contextx.WithUserRole(ctx, s)
			}
		case contextx.FieldIPAddress:
			if s, ok := v.(string); ok {
				ctx = contextx.WithIPAddress(ctx, s)
			}
		case contextx.FieldUserAgent:
			if s, ok := v.(string); ok {
				ctx = contextx.WithUserAgent(ctx, s)
			}
		case contextx.FieldAPIVersion:
			if s, ok := v.(string); ok {
				ctx = contextx.WithAPIVersion(ctx, s)
			}
		}
	}

	return ctx
}

func extractFields(ctx context.Context) map[string]any {
	fields := make(map[string]any)
	if id := contextx.GetRequestID(ctx); id != "" {
		fields[contextx.FieldRequestID] = id
	}
	if id := contextx.GetSessionID(ctx); id != "" {
		fields[contextx.FieldSessionID] = id
	}
	if id := contextx.GetUserID(ctx); id != nil {
		fields[contextx.FieldUserID] = id
	}
	if role := contextx.GetUserRole(ctx); role != "" {
		fields[contextx.FieldUserRole] = role
	}
	if ip := contextx.GetIPAddress(ctx); ip != "" {
		fields[contextx.FieldIPAddress] = ip
	}
	if ua := contextx.GetUserAgent(ctx); ua != "" {
		fields[contextx.FieldUserAgent] = ua
	}
	if v := contextx.GetAPIVersion(ctx); v != "" {
		fields[contextx.FieldAPIVersion] = v
	}
	return fields
}

func mergeFields(fields map[string]any, keysAndValues ...any) []any {
	if len(fields) == 0 {
		return keysAndValues
	}

	return append(keysAndValues, zap.Any("meta_data", fields))
}
