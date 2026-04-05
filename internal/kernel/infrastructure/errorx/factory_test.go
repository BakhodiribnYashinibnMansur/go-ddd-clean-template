package errorx

import (
	"errors"
	"testing"
)

func TestNewBadRequest(t *testing.T) {
	err := NewBadRequest("invalid data")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Type != ErrBadRequest {
		t.Errorf("expected type %s, got %s", ErrBadRequest, err.Type)
	}
	if err.Message != "invalid data" {
		t.Errorf("expected message 'invalid data', got %q", err.Message)
	}
	if err.HTTPStatus != 400 {
		t.Errorf("expected HTTP status 400, got %d", err.HTTPStatus)
	}
}

func TestNewUnauthorized(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantMsg string
	}{
		{
			name:    "with custom message",
			message: "bad credentials",
			wantMsg: "bad credentials",
		},
		{
			name:    "with empty message uses default",
			message: "",
			wantMsg: "Authentication required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewUnauthorized(tt.message)
			if err == nil {
				t.Fatal("expected non-nil error")
			}
			if err.Type != ErrUnauthorized {
				t.Errorf("expected type %s, got %s", ErrUnauthorized, err.Type)
			}
			if err.Message != tt.wantMsg {
				t.Errorf("expected message %q, got %q", tt.wantMsg, err.Message)
			}
			if err.HTTPStatus != 401 {
				t.Errorf("expected HTTP status 401, got %d", err.HTTPStatus)
			}
		})
	}
}

func TestNewForbidden(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantMsg string
	}{
		{
			name:    "with custom message",
			message: "you cannot access this",
			wantMsg: "you cannot access this",
		},
		{
			name:    "with empty message uses default",
			message: "",
			wantMsg: "Access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewForbidden(tt.message)
			if err == nil {
				t.Fatal("expected non-nil error")
			}
			if err.Type != ErrForbidden {
				t.Errorf("expected type %s, got %s", ErrForbidden, err.Type)
			}
			if err.Message != tt.wantMsg {
				t.Errorf("expected message %q, got %q", tt.wantMsg, err.Message)
			}
			if err.HTTPStatus != 403 {
				t.Errorf("expected HTTP status 403, got %d", err.HTTPStatus)
			}
		})
	}
}

func TestNewNotFound(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantMsg string
	}{
		{
			name:    "with custom message",
			message: "user not found",
			wantMsg: "user not found",
		},
		{
			name:    "with empty message uses default",
			message: "",
			wantMsg: "Resource not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewNotFound(tt.message)
			if err == nil {
				t.Fatal("expected non-nil error")
			}
			if err.Type != ErrNotFound {
				t.Errorf("expected type %s, got %s", ErrNotFound, err.Type)
			}
			if err.Message != tt.wantMsg {
				t.Errorf("expected message %q, got %q", tt.wantMsg, err.Message)
			}
			if err.HTTPStatus != 404 {
				t.Errorf("expected HTTP status 404, got %d", err.HTTPStatus)
			}
		})
	}
}

func TestNewConflict(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantMsg string
	}{
		{
			name:    "with custom message",
			message: "duplicate entry",
			wantMsg: "duplicate entry",
		},
		{
			name:    "with empty message uses default",
			message: "",
			wantMsg: "Resource already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewConflict(tt.message)
			if err == nil {
				t.Fatal("expected non-nil error")
			}
			if err.Type != ErrAlreadyExists {
				t.Errorf("expected type %s, got %s", ErrAlreadyExists, err.Type)
			}
			if err.Message != tt.wantMsg {
				t.Errorf("expected message %q, got %q", tt.wantMsg, err.Message)
			}
			if err.HTTPStatus != 409 {
				t.Errorf("expected HTTP status 409, got %d", err.HTTPStatus)
			}
		})
	}
}

func TestNewValidationError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantMsg string
	}{
		{
			name:    "with custom message",
			message: "email is required",
			wantMsg: "email is required",
		},
		{
			name:    "with empty message uses default",
			message: "",
			wantMsg: "Validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewValidationError(tt.message)
			if err == nil {
				t.Fatal("expected non-nil error")
			}
			if err.Type != ErrValidation {
				t.Errorf("expected type %s, got %s", ErrValidation, err.Type)
			}
			if err.Message != tt.wantMsg {
				t.Errorf("expected message %q, got %q", tt.wantMsg, err.Message)
			}
			if err.HTTPStatus != 400 {
				t.Errorf("expected HTTP status 400, got %d", err.HTTPStatus)
			}
		})
	}
}

func TestNewInternalError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantMsg string
	}{
		{
			name:    "with custom message",
			message: "something went wrong",
			wantMsg: "something went wrong",
		},
		{
			name:    "with empty message uses default",
			message: "",
			wantMsg: "An unexpected error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewInternalError(tt.message)
			if err == nil {
				t.Fatal("expected non-nil error")
			}
			if err.Type != ErrInternal {
				t.Errorf("expected type %s, got %s", ErrInternal, err.Type)
			}
			if err.Message != tt.wantMsg {
				t.Errorf("expected message %q, got %q", tt.wantMsg, err.Message)
			}
			if err.HTTPStatus != 500 {
				t.Errorf("expected HTTP status 500, got %d", err.HTTPStatus)
			}
		})
	}
}

func TestNewTimeoutError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantMsg string
	}{
		{
			name:    "with custom message",
			message: "query took too long",
			wantMsg: "query took too long",
		},
		{
			name:    "with empty message uses default",
			message: "",
			wantMsg: "The operation timed out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewTimeoutError(tt.message)
			if err == nil {
				t.Fatal("expected non-nil error")
			}
			if err.Type != ErrTimeout {
				t.Errorf("expected type %s, got %s", ErrTimeout, err.Type)
			}
			if err.Message != tt.wantMsg {
				t.Errorf("expected message %q, got %q", tt.wantMsg, err.Message)
			}
			if err.HTTPStatus != 504 {
				t.Errorf("expected HTTP status 504, got %d", err.HTTPStatus)
			}
		})
	}
}

func TestWrapBadRequest(t *testing.T) {
	origErr := errors.New("original error")
	err := WrapBadRequest(origErr, "bad input")

	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Type != ErrBadRequest {
		t.Errorf("expected type %s, got %s", ErrBadRequest, err.Type)
	}
	if err.Message != "bad input" {
		t.Errorf("expected message 'bad input', got %q", err.Message)
	}
	if err.Err != origErr {
		t.Error("expected wrapped error to be the original error")
	}
}

func TestWrapUnauthorized(t *testing.T) {
	origErr := errors.New("token expired")
	err := WrapUnauthorized(origErr, "auth failed")

	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Type != ErrUnauthorized {
		t.Errorf("expected type %s, got %s", ErrUnauthorized, err.Type)
	}
	if err.Err != origErr {
		t.Error("expected wrapped error to be the original error")
	}
}

func TestWrapForbidden(t *testing.T) {
	origErr := errors.New("no permission")
	err := WrapForbidden(origErr, "access denied")

	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Type != ErrForbidden {
		t.Errorf("expected type %s, got %s", ErrForbidden, err.Type)
	}
}

func TestWrapNotFound(t *testing.T) {
	origErr := errors.New("no rows")
	err := WrapNotFound(origErr, "user missing")

	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Type != ErrNotFound {
		t.Errorf("expected type %s, got %s", ErrNotFound, err.Type)
	}
}

func TestWrapConflict(t *testing.T) {
	origErr := errors.New("duplicate key")
	err := WrapConflict(origErr, "already exists")

	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Type != ErrAlreadyExists {
		t.Errorf("expected type %s, got %s", ErrAlreadyExists, err.Type)
	}
}

func TestWrapInternal(t *testing.T) {
	origErr := errors.New("panic occurred")
	err := WrapInternal(origErr, "internal failure")

	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Type != ErrInternal {
		t.Errorf("expected type %s, got %s", ErrInternal, err.Type)
	}
	if err.Err != origErr {
		t.Error("expected wrapped error to be the original error")
	}
}

func TestWrapWithNilError(t *testing.T) {
	// Wrap functions should return nil when given nil error
	// (based on Wrap function behavior)
	err := WrapBadRequest(nil, "test")
	if err != nil {
		t.Error("expected nil when wrapping nil error")
	}

	err = WrapInternal(nil, "test")
	if err != nil {
		t.Error("expected nil when wrapping nil error")
	}
}
