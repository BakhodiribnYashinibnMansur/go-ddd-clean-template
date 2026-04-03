package errors

import (
	"errors"
	"testing"
)

func TestHandleFirebaseError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		extraFields  map[string]any
		wantNil      bool
		wantType     string
	}{
		{
			name:    "nil error returns nil",
			err:     nil,
			wantNil: true,
		},
		{
			name:     "registration-token-not-registered maps to invalid token",
			err:      errors.New("registration-token-not-registered"),
			wantType: ErrExtFirebaseInvalidToken,
		},
		{
			name:     "invalid-registration-token maps to invalid token",
			err:      errors.New("invalid-registration-token"),
			wantType: ErrExtFirebaseInvalidToken,
		},
		{
			name:     "quota-exceeded maps to quota exceeded",
			err:      errors.New("quota-exceeded"),
			wantType: ErrExtFirebaseQuotaExceeded,
		},
		{
			name:     "message-rate-exceeded maps to quota exceeded",
			err:      errors.New("message-rate-exceeded"),
			wantType: ErrExtFirebaseQuotaExceeded,
		},
		{
			name:     "unavailable maps to unavailable",
			err:      errors.New("unavailable"),
			wantType: ErrExtFirebaseUnavailable,
		},
		{
			name:     "internal-error maps to unavailable",
			err:      errors.New("internal-error"),
			wantType: ErrExtFirebaseUnavailable,
		},
		{
			name:     "timeout maps to unavailable",
			err:      errors.New("timeout"),
			wantType: ErrExtFirebaseUnavailable,
		},
		{
			name:     "deadline maps to unavailable",
			err:      errors.New("deadline"),
			wantType: ErrExtFirebaseUnavailable,
		},
		{
			name:     "unknown error maps to send failed",
			err:      errors.New("something unexpected"),
			wantType: ErrExtFirebaseSendFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HandleFirebaseError(tt.err, tt.extraFields)
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

func TestHandleFirebaseError_ExtraFields(t *testing.T) {
	extra := map[string]any{
		"device_token": "abc123",
		"user_id":      42,
	}

	got := HandleFirebaseError(errors.New("something"), extra)
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
