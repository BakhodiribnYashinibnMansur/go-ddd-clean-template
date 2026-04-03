package errors

import (
	"errors"
	"testing"
)

func TestHandleEventBusError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		channel  string
		wantNil  bool
		wantType string
	}{
		{
			name:    "nil error returns nil",
			err:     nil,
			channel: "events",
			wantNil: true,
		},
		{
			name:     "connection maps to connection error",
			err:      errors.New("connection refused"),
			channel:  "events",
			wantType: ErrExtEventBusConnection,
		},
		{
			name:     "EOF maps to connection error",
			err:      errors.New("unexpected EOF"),
			channel:  "events",
			wantType: ErrExtEventBusConnection,
		},
		{
			name:     "broken pipe maps to connection error",
			err:      errors.New("write: broken pipe"),
			channel:  "events",
			wantType: ErrExtEventBusConnection,
		},
		{
			name:     "timeout maps to timeout",
			err:      errors.New("i/o timeout"),
			channel:  "events",
			wantType: ErrExtEventBusTimeout,
		},
		{
			name:     "deadline maps to timeout",
			err:      errors.New("context deadline exceeded"),
			channel:  "events",
			wantType: ErrExtEventBusTimeout,
		},
		{
			name:     "unknown error maps to publish failed",
			err:      errors.New("something went wrong"),
			channel:  "events",
			wantType: ErrExtEventBusPublishFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HandleEventBusError(tt.err, tt.channel, nil)
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

func TestHandleEventBusError_ChannelField(t *testing.T) {
	got := HandleEventBusError(errors.New("something"), "notifications", nil)
	if got == nil {
		t.Fatal("expected non-nil AppError, got nil")
	}

	v, ok := got.Fields["channel"]
	if !ok {
		t.Fatal("missing 'channel' field")
	}
	if v != "notifications" {
		t.Errorf("Fields[\"channel\"] = %v, want %q", v, "notifications")
	}
}

func TestHandleEventBusError_EmptyChannel(t *testing.T) {
	got := HandleEventBusError(errors.New("something"), "", nil)
	if got == nil {
		t.Fatal("expected non-nil AppError, got nil")
	}

	if got.Fields != nil {
		if _, ok := got.Fields["channel"]; ok {
			t.Error("expected no 'channel' field for empty channel string")
		}
	}
}

func TestHandleEventBusError_ExtraFields(t *testing.T) {
	extra := map[string]any{
		"stream":   "orders",
		"consumer": "worker-1",
	}

	got := HandleEventBusError(errors.New("something"), "orders-channel", extra)
	if got == nil {
		t.Fatal("expected non-nil AppError, got nil")
	}

	// Check channel field
	if v, ok := got.Fields["channel"]; !ok || v != "orders-channel" {
		t.Errorf("Fields[\"channel\"] = %v, want %q", v, "orders-channel")
	}

	// Check extra fields
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
