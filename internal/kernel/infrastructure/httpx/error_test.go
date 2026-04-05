package httpx

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

// mockLogger implements logger.Log and captures Errorw calls.
type mockLogger struct {
	errorwMsg          string
	errorwKeysAndValues []any
}

func (m *mockLogger) Debug(_ ...any)                                {}
func (m *mockLogger) Debugf(_ string, _ ...any)                     {}
func (m *mockLogger) Debugw(_ string, _ ...any)                     {}
func (m *mockLogger) Info(_ ...any)                                 {}
func (m *mockLogger) Infof(_ string, _ ...any)                      {}
func (m *mockLogger) Infow(_ string, _ ...any)                      {}
func (m *mockLogger) Warn(_ ...any)                                 {}
func (m *mockLogger) Warnf(_ string, _ ...any)                      {}
func (m *mockLogger) Warnw(_ string, _ ...any)                      {}
func (m *mockLogger) Error(_ ...any)                                {}
func (m *mockLogger) Errorf(_ string, _ ...any)                     {}
func (m *mockLogger) Fatal(_ ...any)                                {}
func (m *mockLogger) Fatalf(_ string, _ ...any)                     {}
func (m *mockLogger) Fatalw(_ string, _ ...any)                     {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)   {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)   {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any)  {}

func (m *mockLogger) Errorw(msg string, keysAndValues ...any) {
	m.errorwMsg = msg
	m.errorwKeysAndValues = keysAndValues
}

func TestLogError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		msg     string
		wantMsg string
	}{
		{
			name:    "basic error",
			err:     errors.New("something failed"),
			msg:     "handler error",
			wantMsg: "handler error",
		},
		{
			name:    "empty message",
			err:     errors.New("fail"),
			msg:     "",
			wantMsg: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ml := &mockLogger{}
			LogError(ml, tt.err, tt.msg)

			if ml.errorwMsg != tt.wantMsg {
				t.Errorf("Errorw msg = %q, want %q", ml.errorwMsg, tt.wantMsg)
			}
			if len(ml.errorwKeysAndValues) < 2 {
				t.Fatalf("Errorw keysAndValues has %d elements, want at least 2", len(ml.errorwKeysAndValues))
			}
		})
	}
}

func TestGlobalErrorVariables(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrParamIsEmpty", ErrParamIsEmpty},
		{"ErrParsingQuery", ErrParsingQuery},
		{"ErrUnmarshalData", ErrUnmarshalData},
		{"ErrParamIsInvalid", ErrParamIsInvalid},
		{"ErrParsingUUID", ErrParsingUUID},
		{"ErrUnAuth", ErrUnAuth},
		{"ErrInvalidToken", ErrInvalidToken},
		{"ErrExpiredToken", ErrExpiredToken},
		{"ErrRoleNotFound", ErrRoleNotFound},
		{"ErrAccessDenied", ErrAccessDenied},
		{"ErrRateLimitExceeded", ErrRateLimitExceeded},
		{"ErrNotImplemented", ErrNotImplemented},
		{"ErrStorageNotConfigured", ErrStorageNotConfigured},
		{"ErrInternalError", ErrInternalError},
		{"ErrPanicRecovered", ErrPanicRecovered},
		{"ErrFileRequired", ErrFileRequired},
		{"ErrFileNotFound", ErrFileNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s is nil, want non-nil error", tt.name)
			}
		})
	}
}

func TestFormatConstants(t *testing.T) {
	tests := []struct {
		name   string
		format string
		arg    string
		want   string
	}{
		{
			name:   "ParamInvalid",
			format: ParamInvalid,
			arg:    "id",
			want:   "parameter id is invalid",
		},
		{
			name:   "QueryInvalid",
			format: QueryInvalid,
			arg:    "page",
			want:   "query parameter page is invalid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fmt.Sprintf(tt.format, tt.arg)
			if got != tt.want {
				t.Errorf("fmt.Sprintf(%q, %q) = %q, want %q", tt.format, tt.arg, got, tt.want)
			}
		})
	}
}
