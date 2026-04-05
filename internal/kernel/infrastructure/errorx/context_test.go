package errorx

import (
	"context"
	"strings"
	"testing"
)

func TestGetErrorContext_NilContext(t *testing.T) {
	ec := GetErrorContext(nil)
	if ec == nil {
		t.Fatal("expected non-nil ErrorContext for nil context")
	}
	if ec.Metadata == nil {
		t.Error("expected initialized Metadata map")
	}
	if ec.UserID != "" {
		t.Error("expected empty UserID")
	}
}

func TestGetErrorContext_EmptyContext(t *testing.T) {
	ctx := context.Background()
	ec := GetErrorContext(ctx)
	if ec == nil {
		t.Fatal("expected non-nil ErrorContext")
	}
	if ec.Metadata == nil {
		t.Error("expected initialized Metadata map")
	}
}

func TestGetErrorContext_WithExistingContext(t *testing.T) {
	expected := &ErrorContext{
		UserID:    "user-123",
		RequestID: "req-456",
		Metadata:  map[string]any{"key": "value"},
	}

	ctx := WithErrorContext(context.Background(), expected)
	got := GetErrorContext(ctx)

	if got.UserID != expected.UserID {
		t.Errorf("expected UserID %q, got %q", expected.UserID, got.UserID)
	}
	if got.RequestID != expected.RequestID {
		t.Errorf("expected RequestID %q, got %q", expected.RequestID, got.RequestID)
	}
}

func TestGetErrorContext_FromIndividualValues(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, ContextKeyUserID, "user-abc")
	ctx = context.WithValue(ctx, ContextKeyRequestID, "req-def")

	ec := GetErrorContext(ctx)

	if ec.UserID != "user-abc" {
		t.Errorf("expected UserID 'user-abc', got %q", ec.UserID)
	}
	if ec.RequestID != "req-def" {
		t.Errorf("expected RequestID 'req-def', got %q", ec.RequestID)
	}
}

func TestWithErrorContext(t *testing.T) {
	ec := &ErrorContext{
		UserID:    "u1",
		Operation: "create",
	}
	ctx := WithErrorContext(context.Background(), ec)

	got := ctx.Value(ContextKeyErrorContext)
	if got == nil {
		t.Fatal("expected ErrorContext in context")
	}
	gotEC, ok := got.(*ErrorContext)
	if !ok {
		t.Fatal("expected *ErrorContext type")
	}
	if gotEC.UserID != "u1" {
		t.Errorf("expected UserID 'u1', got %q", gotEC.UserID)
	}
}

func TestWithSource(t *testing.T) {
	t.Run("with valid error", func(t *testing.T) {
		err := New(ErrInternal, "test")
		result := WithSource(err, "repo/user.go", "UserRepo.Get")

		if result == nil {
			t.Fatal("expected non-nil error")
		}
		if result.Fields["file"] != "repo/user.go" {
			t.Errorf("expected file 'repo/user.go', got %v", result.Fields["file"])
		}
		if result.Fields["function"] != "UserRepo.Get" {
			t.Errorf("expected function 'UserRepo.Get', got %v", result.Fields["function"])
		}
	})

	t.Run("with nil error", func(t *testing.T) {
		result := WithSource(nil, "file.go", "func")
		if result != nil {
			t.Error("expected nil for nil error input")
		}
	})
}

func TestGetCaller(t *testing.T) {
	file, function := GetCaller(0)

	if file == unknownValue {
		t.Error("expected non-unknown file")
	}
	if function == unknownValue {
		t.Error("expected non-unknown function")
	}
	// Should contain the test file path
	if !strings.Contains(file, "context_test.go") {
		t.Errorf("expected file to contain 'context_test.go', got %q", file)
	}
}

func TestWithCaller(t *testing.T) {
	t.Run("with valid error", func(t *testing.T) {
		err := New(ErrInternal, "test")
		result := WithCaller(err, 0)

		if result == nil {
			t.Fatal("expected non-nil error")
		}
		if result.Fields["file"] == nil {
			t.Error("expected file field to be set")
		}
		if result.Fields["function"] == nil {
			t.Error("expected function field to be set")
		}
	})

	t.Run("with nil error", func(t *testing.T) {
		result := WithCaller(nil, 0)
		if result != nil {
			t.Error("expected nil for nil error input")
		}
	})
}

func TestAutoSource(t *testing.T) {
	t.Run("with valid error", func(t *testing.T) {
		err := New(ErrInternal, "test")
		result := AutoSource(err)

		if result == nil {
			t.Fatal("expected non-nil error")
		}
		if result.Fields["file"] == nil {
			t.Error("expected file field to be set")
		}
		if result.Fields["function"] == nil {
			t.Error("expected function field to be set")
		}
	})

	t.Run("with nil error", func(t *testing.T) {
		result := AutoSource(nil)
		if result != nil {
			t.Error("expected nil for nil error input")
		}
	})
}

func TestAppError_WithOperation(t *testing.T) {
	err := New(ErrInternal, "test")
	result := err.WithOperation("create_user")

	if result.Fields["operation"] != "create_user" {
		t.Errorf("expected operation 'create_user', got %v", result.Fields["operation"])
	}
}

func TestAppError_WithResource(t *testing.T) {
	t.Run("with resource and ID", func(t *testing.T) {
		err := New(ErrNotFound, "test")
		result := err.WithResource("user", "123")

		if result.Fields["resource"] != "user" {
			t.Errorf("expected resource 'user', got %v", result.Fields["resource"])
		}
		if result.Fields["resource_id"] != "123" {
			t.Errorf("expected resource_id '123', got %v", result.Fields["resource_id"])
		}
	})

	t.Run("with resource and empty ID", func(t *testing.T) {
		err := New(ErrNotFound, "test")
		result := err.WithResource("user", "")

		if result.Fields["resource"] != "user" {
			t.Errorf("expected resource 'user', got %v", result.Fields["resource"])
		}
		if _, ok := result.Fields["resource_id"]; ok {
			t.Error("expected resource_id to not be set for empty ID")
		}
	})
}

func TestAppError_WithContext(t *testing.T) {
	ec := &ErrorContext{
		UserID:    "user-1",
		RequestID: "req-1",
		Operation: "get_user",
		Resource:  "user",
		IPAddress: "127.0.0.1",
		UserAgent: "TestAgent",
		Path:      "/api/users",
		Method:    "GET",
		Metadata:  map[string]any{"extra": "data"},
	}
	ctx := WithErrorContext(context.Background(), ec)

	err := New(ErrInternal, "test")
	result := err.WithContext(ctx)

	checks := map[string]any{
		"user_id":    "user-1",
		"request_id": "req-1",
		"operation":  "get_user",
		"resource":   "user",
		"ip_address": "127.0.0.1",
		"user_agent": "TestAgent",
		"path":       "/api/users",
		"method":     "GET",
		"extra":      "data",
	}

	for key, want := range checks {
		got, ok := result.Fields[key]
		if !ok {
			t.Errorf("expected field %q to be set", key)
			continue
		}
		if got != want {
			t.Errorf("field %q: expected %v, got %v", key, want, got)
		}
	}
}

func TestAppError_WithMetadata(t *testing.T) {
	err := New(ErrInternal, "test")
	result := err.WithMetadata("attempt", 3)

	if result.Fields["attempt"] != 3 {
		t.Errorf("expected attempt 3, got %v", result.Fields["attempt"])
	}
}

func TestAppError_WithTag(t *testing.T) {
	err := New(ErrInternal, "test")
	result := err.WithTag("database").WithTag("critical")

	tags, ok := result.Fields["tags"].([]string)
	if !ok {
		t.Fatal("expected tags to be []string")
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	if tags[0] != "database" {
		t.Errorf("expected first tag 'database', got %q", tags[0])
	}
	if tags[1] != "critical" {
		t.Errorf("expected second tag 'critical', got %q", tags[1])
	}
}

func TestAppError_GetMetadata(t *testing.T) {
	err := New(ErrBadRequest, "test")
	meta := err.GetMetadata()

	if meta.Category != CategoryValidation {
		t.Errorf("expected category %s, got %s", CategoryValidation, meta.Category)
	}
	if meta.Severity != SeverityLow {
		t.Errorf("expected severity %s, got %s", SeverityLow, meta.Severity)
	}
}

func TestAppError_IsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		errType  string
		expected bool
	}{
		{"timeout is retryable", ErrTimeout, true},
		{"repo timeout is retryable", ErrRepoTimeout, true},
		{"repo connection is retryable", ErrRepoConnection, true},
		{"bad request is not retryable", ErrBadRequest, false},
		{"not found is not retryable", ErrNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.errType, "test")
			if got := err.IsRetryable(); got != tt.expected {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAppError_GetSeverity(t *testing.T) {
	err := New(ErrRepoDatabase, "test")
	if got := err.GetSeverity(); got != SeverityCritical {
		t.Errorf("expected severity %s, got %s", SeverityCritical, got)
	}
}

func TestAppError_GetCategory(t *testing.T) {
	err := New(ErrBadRequest, "test")
	if got := err.GetCategory(); got != CategoryValidation {
		t.Errorf("expected category %s, got %s", CategoryValidation, got)
	}
}

func TestAppError_String(t *testing.T) {
	err := New(ErrBadRequest, "invalid input")
	s := err.String()

	if !strings.Contains(s, "BAD_REQUEST") {
		t.Errorf("expected String() to contain 'BAD_REQUEST', got %q", s)
	}
	if !strings.Contains(s, "invalid input") {
		t.Errorf("expected String() to contain 'invalid input', got %q", s)
	}
}
