package logger

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestIsSensitive(t *testing.T) {
	t.Helper()

	sensitiveList := []string{
		"password", "token", "access_token", "refresh_token",
		"api_key", "secret", "authorization", "cookie",
		"csrf_token", "otp", "pin",
	}

	t.Run("all sensitive keys detected", func(t *testing.T) {
		t.Helper()
		for _, key := range sensitiveList {
			if !isSensitive(key) {
				t.Errorf("expected %q to be sensitive", key)
			}
		}
	})

	t.Run("non-sensitive keys pass through", func(t *testing.T) {
		t.Helper()
		nonSensitive := []string{"username", "email", "name", "status", "id"}
		for _, key := range nonSensitive {
			if isSensitive(key) {
				t.Errorf("expected %q to not be sensitive", key)
			}
		}
	})

	t.Run("case insensitivity", func(t *testing.T) {
		t.Helper()
		cases := []string{"Password", "PASSWORD", "Token", "API_KEY", "Secret"}
		for _, key := range cases {
			if !isSensitive(key) {
				t.Errorf("expected %q to be sensitive (case insensitive)", key)
			}
		}
	})
}

func TestRedactFields(t *testing.T) {
	t.Helper()

	tests := []struct {
		name   string
		fields []zapcore.Field
		want   map[string]string // key -> expected string value ("***" if redacted)
	}{
		{
			name:   "empty fields slice",
			fields: []zapcore.Field{},
			want:   map[string]string{},
		},
		{
			name: "sensitive field redacted",
			fields: []zapcore.Field{
				zapcore.Field{Key: "password", Type: zapcore.StringType, String: "s3cret"},
			},
			want: map[string]string{"password": redactedValue},
		},
		{
			name: "non-sensitive field unchanged",
			fields: []zapcore.Field{
				zapcore.Field{Key: "username", Type: zapcore.StringType, String: "alice"},
			},
			want: map[string]string{"username": "alice"},
		},
		{
			name: "mixed sensitive and non-sensitive",
			fields: []zapcore.Field{
				{Key: "username", Type: zapcore.StringType, String: "bob"},
				{Key: "token", Type: zapcore.StringType, String: "abc123"},
				{Key: "status", Type: zapcore.StringType, String: "active"},
				{Key: "api_key", Type: zapcore.StringType, String: "key-xyz"},
			},
			want: map[string]string{
				"username": "bob",
				"token":    redactedValue,
				"status":   "active",
				"api_key":  redactedValue,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			got := redactFields(tt.fields)
			if len(got) != len(tt.fields) {
				t.Fatalf("len(got) = %d, want %d", len(got), len(tt.fields))
			}
			for i, f := range got {
				wantStr, ok := tt.want[f.Key]
				if !ok {
					t.Fatalf("unexpected key %q at index %d", f.Key, i)
				}
				if f.String != wantStr {
					t.Errorf("field %q: got %q, want %q", f.Key, f.String, wantStr)
				}
				// Redacted fields must have StringType
				if wantStr == redactedValue && f.Type != zapcore.StringType {
					t.Errorf("field %q: expected StringType for redacted field, got %v", f.Key, f.Type)
				}
			}
		})
	}
}

// mockCore is a minimal zapcore.Core for testing redactCore wrapping behaviour.
type mockCore struct {
	enabled     bool
	level       zapcore.Level
	withFields  []zapcore.Field
	writeEntry  zapcore.Entry
	writeFields []zapcore.Field
	writeCalled bool
}

func (c *mockCore) Enabled(l zapcore.Level) bool { return l >= c.level }

func (c *mockCore) With(fields []zapcore.Field) zapcore.Core {
	return &mockCore{
		enabled:    c.enabled,
		level:      c.level,
		withFields: fields,
	}
}

func (c *mockCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *mockCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	c.writeCalled = true
	c.writeEntry = ent
	c.writeFields = fields
	return nil
}

func (c *mockCore) Sync() error { return nil }

func TestRedactCore_With(t *testing.T) {
	t.Helper()

	inner := &mockCore{level: zapcore.DebugLevel}
	rc := NewRedactCore(inner)

	fields := []zapcore.Field{
		{Key: "password", Type: zapcore.StringType, String: "secret123"},
		{Key: "name", Type: zapcore.StringType, String: "alice"},
	}

	newCore := rc.(*redactCore).With(fields)
	innerNew := newCore.(*redactCore).Core.(*mockCore)

	if len(innerNew.withFields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(innerNew.withFields))
	}
	if innerNew.withFields[0].String != redactedValue {
		t.Errorf("password field not redacted: got %q", innerNew.withFields[0].String)
	}
	if innerNew.withFields[1].String != "alice" {
		t.Errorf("name field changed: got %q", innerNew.withFields[1].String)
	}
}

func TestRedactCore_Write(t *testing.T) {
	t.Helper()

	inner := &mockCore{level: zapcore.DebugLevel}
	rc := &redactCore{Core: inner}

	ent := zapcore.Entry{Level: zapcore.InfoLevel, Message: "test"}
	fields := []zapcore.Field{
		{Key: "token", Type: zapcore.StringType, String: "tok-abc"},
	}

	err := rc.Write(ent, fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !inner.writeCalled {
		t.Fatal("inner Write was not called")
	}
	if inner.writeFields[0].String != redactedValue {
		t.Errorf("token not redacted in Write: got %q", inner.writeFields[0].String)
	}
}

func TestRedactCore_Check(t *testing.T) {
	t.Helper()

	t.Run("entry level above core level - core is enabled", func(t *testing.T) {
		t.Helper()

		inner := &mockCore{level: zapcore.DebugLevel}
		rc := &redactCore{Core: inner}

		ent := zapcore.Entry{Level: zapcore.InfoLevel}
		if !rc.Enabled(ent.Level) {
			t.Error("expected core to be enabled for InfoLevel")
		}
	})

	t.Run("entry level below core level - core is disabled", func(t *testing.T) {
		t.Helper()

		inner := &mockCore{level: zapcore.ErrorLevel}
		rc := &redactCore{Core: inner}

		ent := zapcore.Entry{Level: zapcore.DebugLevel}
		if rc.Enabled(ent.Level) {
			t.Error("expected core to be disabled for DebugLevel")
		}
	})
}
