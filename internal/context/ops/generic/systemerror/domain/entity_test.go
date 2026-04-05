package domain_test

import (
	"testing"
	"time"

	domain "gct/internal/context/ops/generic/systemerror/domain"

	"github.com/google/uuid"
)

func se_now() time.Time { return time.Now() }

func TestNewSystemError(t *testing.T) {
	t.Parallel()

	se := domain.NewSystemError("INTERNAL_ERROR", "something went wrong", "ERROR")

	if se.Code() != "INTERNAL_ERROR" {
		t.Fatalf("expected code INTERNAL_ERROR, got %s", se.Code())
	}
	if se.Message() != "something went wrong" {
		t.Fatalf("expected message 'something went wrong', got %s", se.Message())
	}
	if se.Severity() != "ERROR" {
		t.Fatalf("expected severity ERROR, got %s", se.Severity())
	}
	if se.IsResolved() {
		t.Fatal("new system error should not be resolved")
	}
	if se.ResolvedAt() != nil {
		t.Fatal("resolvedAt should be nil")
	}
	if se.ResolvedBy() != nil {
		t.Fatal("resolvedBy should be nil")
	}
	if se.StackTrace() != nil {
		t.Fatal("stackTrace should be nil by default")
	}
	if se.Metadata() == nil {
		t.Fatal("metadata should be initialized (not nil)")
	}
	if se.ServiceName() != nil {
		t.Fatal("serviceName should be nil by default")
	}

	// Should have a SystemErrorRecorded event.
	events := se.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventName() != "system_error.recorded" {
		t.Fatalf("expected system_error.recorded, got %s", events[0].EventName())
	}
}

func TestSystemError_Resolve(t *testing.T) {
	t.Parallel()

	se := domain.NewSystemError("INTERNAL_ERROR", "something went wrong", "ERROR")
	se.ClearEvents()

	resolverID := uuid.New()
	se.Resolve(resolverID)

	if !se.IsResolved() {
		t.Fatal("system error should be resolved after Resolve()")
	}
	if se.ResolvedAt() == nil {
		t.Fatal("resolvedAt should be set after Resolve()")
	}
	if se.ResolvedBy() == nil || *se.ResolvedBy() != resolverID {
		t.Fatal("resolvedBy should match the resolver ID")
	}

	// Should have a SystemErrorResolved event.
	events := se.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventName() != "system_error.resolved" {
		t.Fatalf("expected system_error.resolved, got %s", events[0].EventName())
	}
}

func TestNewSystemError_Setters(t *testing.T) {
	t.Parallel()

	se := domain.NewSystemError("CODE", "msg", "WARN")

	st := "stack trace here"
	se.SetStackTrace(&st)
	if se.StackTrace() == nil || *se.StackTrace() != st {
		t.Fatal("stack trace not set correctly")
	}

	meta := map[string]string{"key": "value"}
	se.SetMetadata(meta)
	if se.Metadata()["key"] != "value" {
		t.Fatal("metadata not set correctly")
	}

	svc := "my-service"
	se.SetServiceName(&svc)
	if se.ServiceName() == nil || *se.ServiceName() != svc {
		t.Fatal("serviceName not set correctly")
	}

	reqID := uuid.New()
	se.SetRequestID(&reqID)
	if se.RequestID() == nil || *se.RequestID() != reqID {
		t.Fatal("requestID not set correctly")
	}

	userID := uuid.New()
	se.SetUserID(&userID)
	if se.UserID() == nil || *se.UserID() != userID {
		t.Fatal("userID not set correctly")
	}

	ip := "127.0.0.1"
	se.SetIPAddress(&ip)
	if se.IPAddress() == nil || *se.IPAddress() != ip {
		t.Fatal("ipAddress not set correctly")
	}

	path := "/api/test"
	se.SetPath(&path)
	if se.Path() == nil || *se.Path() != path {
		t.Fatal("path not set correctly")
	}

	method := "GET"
	se.SetMethod(&method)
	if se.Method() == nil || *se.Method() != method {
		t.Fatal("method not set correctly")
	}
}

func TestReconstructSystemError(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	code := "DB_ERROR"
	msg := "connection failed"
	severity := "FATAL"
	stack := "goroutine 1..."
	meta := map[string]string{"db": "postgres"}

	se := domain.ReconstructSystemError(
		id, se_now(),
		code, msg, &stack, meta, severity,
		nil, nil, nil, nil, nil, nil,
		false, nil, nil,
	)

	if se.ID() != id {
		t.Fatal("ID mismatch")
	}
	if se.Code() != code {
		t.Fatal("code mismatch")
	}
	if se.Message() != msg {
		t.Fatal("message mismatch")
	}
	if se.Severity() != severity {
		t.Fatal("severity mismatch")
	}
	if se.StackTrace() == nil || *se.StackTrace() != stack {
		t.Fatal("stackTrace mismatch")
	}
	if se.Metadata()["db"] != "postgres" {
		t.Fatal("metadata mismatch")
	}

	// Reconstruct should not raise events.
	if len(se.Events()) != 0 {
		t.Fatalf("expected 0 events, got %d", len(se.Events()))
	}
}
