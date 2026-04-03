package logger

import (
	"context"
	"testing"

	"gct/internal/shared/infrastructure/contextx"
)

func TestExtractFields(t *testing.T) {
	t.Helper()

	tests := []struct {
		name       string
		buildCtx   func() context.Context
		wantKeys   []string
		wantValues map[string]any
	}{
		{
			name: "all context values set",
			buildCtx: func() context.Context {
				ctx := context.Background()
				ctx = contextx.WithRequestID(ctx, "req-1")
				ctx = contextx.WithSessionID(ctx, "sess-1")
				ctx = contextx.WithUserID(ctx, "user-1")
				ctx = contextx.WithUserRole(ctx, "admin")
				ctx = contextx.WithIPAddress(ctx, "10.0.0.1")
				ctx = contextx.WithUserAgent(ctx, "TestAgent/1.0")
				ctx = contextx.WithAPIVersion(ctx, "v2")
				return ctx
			},
			wantKeys: []string{
				contextx.FieldRequestID, contextx.FieldSessionID,
				contextx.FieldUserID, contextx.FieldUserRole,
				contextx.FieldIPAddress, contextx.FieldUserAgent,
				contextx.FieldAPIVersion,
			},
			wantValues: map[string]any{
				contextx.FieldRequestID:  "req-1",
				contextx.FieldSessionID:  "sess-1",
				contextx.FieldUserID:     "user-1",
				contextx.FieldUserRole:   "admin",
				contextx.FieldIPAddress:  "10.0.0.1",
				contextx.FieldUserAgent:  "TestAgent/1.0",
				contextx.FieldAPIVersion: "v2",
			},
		},
		{
			name: "empty context",
			buildCtx: func() context.Context {
				return context.Background()
			},
			wantKeys:   nil,
			wantValues: map[string]any{},
		},
		{
			name: "partial context - only request_id and user_role",
			buildCtx: func() context.Context {
				ctx := context.Background()
				ctx = contextx.WithRequestID(ctx, "req-partial")
				ctx = contextx.WithUserRole(ctx, "viewer")
				return ctx
			},
			wantKeys: []string{contextx.FieldRequestID, contextx.FieldUserRole},
			wantValues: map[string]any{
				contextx.FieldRequestID: "req-partial",
				contextx.FieldUserRole:  "viewer",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			ctx := tt.buildCtx()
			fields := extractFields(ctx)

			if tt.wantKeys == nil {
				if len(fields) != 0 {
					t.Fatalf("expected empty fields, got %v", fields)
				}
				return
			}

			if len(fields) != len(tt.wantKeys) {
				t.Fatalf("field count = %d, want %d; fields: %v", len(fields), len(tt.wantKeys), fields)
			}

			for _, key := range tt.wantKeys {
				got, ok := fields[key]
				if !ok {
					t.Errorf("missing key %q", key)
					continue
				}
				want := tt.wantValues[key]
				if got != want {
					t.Errorf("key %q: got %v, want %v", key, got, want)
				}
			}
		})
	}
}

func TestMergeFields(t *testing.T) {
	t.Helper()

	tests := []struct {
		name          string
		fields        map[string]any
		keysAndValues []any
		wantLen       int
		checkMetaData bool
	}{
		{
			name:          "empty fields map returns keysAndValues as-is",
			fields:        map[string]any{},
			keysAndValues: []any{"key1", "val1"},
			wantLen:       2,
			checkMetaData: false,
		},
		{
			name:          "nil fields map returns keysAndValues as-is",
			fields:        nil,
			keysAndValues: []any{"key1", "val1"},
			wantLen:       2,
			checkMetaData: false,
		},
		{
			name:          "non-empty fields appends meta_data",
			fields:        map[string]any{"request_id": "req-1"},
			keysAndValues: []any{"op", "test"},
			wantLen:       3, // "op", "test", zap.Any("meta_data", ...) as single Field element
			checkMetaData: false,
		},
		{
			name:          "non-empty fields with empty keysAndValues",
			fields:        map[string]any{"user_id": "u1"},
			keysAndValues: nil,
			wantLen:       1, // zap.Any("meta_data", ...) as single Field element
			checkMetaData: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			got := mergeFields(tt.fields, tt.keysAndValues...)

			if len(got) != tt.wantLen {
				t.Fatalf("len = %d, want %d; got: %v", len(got), tt.wantLen, got)
			}

			if tt.checkMetaData {
				metaKey, ok := got[len(got)-2].(string)
				if !ok || metaKey != "meta_data" {
					t.Errorf("expected 'meta_data' key, got %v", got[len(got)-2])
				}
			}
		})
	}
}

func TestWithFields(t *testing.T) {
	t.Helper()

	tests := []struct {
		name   string
		fields map[string]any
		check  func(t *testing.T, ctx context.Context)
	}{
		{
			name: "sets all supported field types",
			fields: map[string]any{
				contextx.FieldRequestID:  "req-wf",
				contextx.FieldSessionID:  "sess-wf",
				contextx.FieldUserID:     "uid-wf",
				contextx.FieldUserRole:   "editor",
				contextx.FieldIPAddress:  "192.168.1.1",
				contextx.FieldUserAgent:  "Mozilla/5.0",
				contextx.FieldAPIVersion: "v3",
			},
			check: func(t *testing.T, ctx context.Context) {
				t.Helper()
				utilAssertEqual(t, "request_id", contextx.GetRequestID(ctx), "req-wf")
				utilAssertEqual(t, "session_id", contextx.GetSessionID(ctx), "sess-wf")
				if contextx.GetUserID(ctx) != "uid-wf" {
					t.Errorf("user_id: got %v, want %v", contextx.GetUserID(ctx), "uid-wf")
				}
				utilAssertEqual(t, "user_role", contextx.GetUserRole(ctx), "editor")
				utilAssertEqual(t, "ip_address", contextx.GetIPAddress(ctx), "192.168.1.1")
				utilAssertEqual(t, "user_agent", contextx.GetUserAgent(ctx), "Mozilla/5.0")
				utilAssertEqual(t, "api_version", contextx.GetAPIVersion(ctx), "v3")
			},
		},
		{
			name: "ignores unknown keys",
			fields: map[string]any{
				"unknown_field":         "ignored",
				"another_unknown":       123,
				contextx.FieldRequestID: "req-ok",
			},
			check: func(t *testing.T, ctx context.Context) {
				t.Helper()
				utilAssertEqual(t, "request_id", contextx.GetRequestID(ctx), "req-ok")
			},
		},
		{
			name:   "empty fields map",
			fields: map[string]any{},
			check: func(t *testing.T, ctx context.Context) {
				t.Helper()
				if contextx.GetRequestID(ctx) != "" {
					t.Error("expected empty request_id")
				}
			},
		},
		{
			name: "wrong type for string field is ignored",
			fields: map[string]any{
				contextx.FieldRequestID: 12345,
			},
			check: func(t *testing.T, ctx context.Context) {
				t.Helper()
				if got := contextx.GetRequestID(ctx); got != "" {
					t.Errorf("expected empty request_id for wrong type, got %q", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			ctx := context.Background()
			ctx = WithFields(ctx, tt.fields)
			tt.check(t, ctx)
		})
	}
}

func utilAssertEqual(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %q, want %q", field, got, want)
	}
}
