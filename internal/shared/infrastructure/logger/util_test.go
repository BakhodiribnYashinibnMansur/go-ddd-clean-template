package logger

import (
	"context"
	"testing"

	"gct/internal/shared/infrastructure/contextx"
)

func TestWithFields_RequestID(t *testing.T) {
	ctx := context.Background()
	ctx = WithFields(ctx, map[string]any{
		contextx.FieldRequestID: "req-123",
	})

	if got := contextx.GetRequestID(ctx); got != "req-123" {
		t.Errorf("expected request ID 'req-123', got %q", got)
	}
}

func TestWithFields_SessionID(t *testing.T) {
	ctx := context.Background()
	ctx = WithFields(ctx, map[string]any{
		contextx.FieldSessionID: "sess-456",
	})

	if got := contextx.GetSessionID(ctx); got != "sess-456" {
		t.Errorf("expected session ID 'sess-456', got %q", got)
	}
}

func TestWithFields_UserID(t *testing.T) {
	ctx := context.Background()
	ctx = WithFields(ctx, map[string]any{
		contextx.FieldUserID: "user-789",
	})

	if got := contextx.GetUserID(ctx); got != "user-789" {
		t.Errorf("expected user ID 'user-789', got %v", got)
	}
}

func TestWithFields_UserRole(t *testing.T) {
	ctx := context.Background()
	ctx = WithFields(ctx, map[string]any{
		contextx.FieldUserRole: "admin",
	})

	if got := contextx.GetUserRole(ctx); got != "admin" {
		t.Errorf("expected user role 'admin', got %q", got)
	}
}

func TestWithFields_IPAddress(t *testing.T) {
	ctx := context.Background()
	ctx = WithFields(ctx, map[string]any{
		contextx.FieldIPAddress: "192.168.1.1",
	})

	if got := contextx.GetIPAddress(ctx); got != "192.168.1.1" {
		t.Errorf("expected IP '192.168.1.1', got %q", got)
	}
}

func TestWithFields_UserAgent(t *testing.T) {
	ctx := context.Background()
	ctx = WithFields(ctx, map[string]any{
		contextx.FieldUserAgent: "TestAgent/1.0",
	})

	if got := contextx.GetUserAgent(ctx); got != "TestAgent/1.0" {
		t.Errorf("expected user agent 'TestAgent/1.0', got %q", got)
	}
}

func TestWithFields_APIVersion(t *testing.T) {
	ctx := context.Background()
	ctx = WithFields(ctx, map[string]any{
		contextx.FieldAPIVersion: "v2",
	})

	if got := contextx.GetAPIVersion(ctx); got != "v2" {
		t.Errorf("expected API version 'v2', got %q", got)
	}
}

func TestWithFields_MultipleFields(t *testing.T) {
	ctx := context.Background()
	ctx = WithFields(ctx, map[string]any{
		contextx.FieldRequestID: "req-1",
		contextx.FieldSessionID: "sess-2",
		contextx.FieldUserRole:  "editor",
	})

	if got := contextx.GetRequestID(ctx); got != "req-1" {
		t.Errorf("expected request ID 'req-1', got %q", got)
	}
	if got := contextx.GetSessionID(ctx); got != "sess-2" {
		t.Errorf("expected session ID 'sess-2', got %q", got)
	}
	if got := contextx.GetUserRole(ctx); got != "editor" {
		t.Errorf("expected user role 'editor', got %q", got)
	}
}

func TestWithFields_IgnoresWrongType(t *testing.T) {
	ctx := context.Background()
	// Pass an int instead of string for RequestID - should be ignored
	ctx = WithFields(ctx, map[string]any{
		contextx.FieldRequestID: 12345,
	})

	if got := contextx.GetRequestID(ctx); got != "" {
		t.Errorf("expected empty request ID for wrong type, got %q", got)
	}
}

func TestExtractFields_Empty(t *testing.T) {
	ctx := context.Background()
	fields := extractFields(ctx)
	if len(fields) != 0 {
		t.Errorf("expected 0 fields for empty context, got %d", len(fields))
	}
}

func TestExtractFields_WithValues(t *testing.T) {
	ctx := context.Background()
	ctx = contextx.WithRequestID(ctx, "req-1")
	ctx = contextx.WithSessionID(ctx, "sess-2")
	ctx = contextx.WithUserRole(ctx, "admin")

	fields := extractFields(ctx)
	if fields[contextx.FieldRequestID] != "req-1" {
		t.Errorf("expected request_id 'req-1', got %v", fields[contextx.FieldRequestID])
	}
	if fields[contextx.FieldSessionID] != "sess-2" {
		t.Errorf("expected session_id 'sess-2', got %v", fields[contextx.FieldSessionID])
	}
	if fields[contextx.FieldUserRole] != "admin" {
		t.Errorf("expected user_role 'admin', got %v", fields[contextx.FieldUserRole])
	}
}

func TestMergeFields_Empty(t *testing.T) {
	fields := map[string]any{}
	result := mergeFields(fields, "key1", "val1")
	if len(result) != 2 {
		t.Errorf("expected 2 elements, got %d", len(result))
	}
}

func TestMergeFields_WithFields(t *testing.T) {
	fields := map[string]any{"request_id": "req-1"}
	result := mergeFields(fields, "key1", "val1")
	// Should have key1, val1, and the zap.Field for meta_data (single element from zap.Any)
	if len(result) < 3 {
		t.Errorf("expected at least 3 elements (key1, val1, meta_data zap.Field), got %d", len(result))
	}
}
