package errors

import (
	"errors"
	"testing"
)

func TestHandleTelegramError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantNil  bool
		wantType string
	}{
		{
			name:    "nil error returns nil",
			err:     nil,
			wantNil: true,
		},
		{
			name:     "429 maps to rate limit",
			err:      errors.New("HTTP 429 response"),
			wantType: ErrExtTelegramRateLimit,
		},
		{
			name:     "Too Many Requests maps to rate limit",
			err:      errors.New("Too Many Requests"),
			wantType: ErrExtTelegramRateLimit,
		},
		{
			name:     "timeout maps to timeout",
			err:      errors.New("timeout waiting for response"),
			wantType: ErrExtTelegramTimeout,
		},
		{
			name:     "deadline maps to timeout",
			err:      errors.New("context deadline exceeded"),
			wantType: ErrExtTelegramTimeout,
		},
		{
			name:     "connection maps to connection error",
			err:      errors.New("connection refused"),
			wantType: ErrExtTelegramConnection,
		},
		{
			name:     "dial maps to connection error",
			err:      errors.New("dial tcp: lookup failed"),
			wantType: ErrExtTelegramConnection,
		},
		{
			name:     "unknown error maps to API error",
			err:      errors.New("unexpected response"),
			wantType: ErrExtTelegramAPIError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HandleTelegramError(tt.err, nil)
			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil AppError, got nil")
			}
			if got.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", got.Type, tt.wantType)
			}
		})
	}
}

func TestHandleTelegramError_ExtraFields(t *testing.T) {
	extra := map[string]any{
		"chat_id":    int64(12345),
		"message_id": "msg-001",
	}

	got := HandleTelegramError(errors.New("unexpected response"), extra)
	if got == nil {
		t.Fatal("expected non-nil AppError, got nil")
	}

	for k, want := range extra {
		v, ok := got.Fields[k]
		if !ok {
			t.Errorf("missing field %q", k)
			continue
		}
		if v != want {
			t.Errorf("Fields[%q] = %v, want %v", k, v, want)
		}
	}
}
