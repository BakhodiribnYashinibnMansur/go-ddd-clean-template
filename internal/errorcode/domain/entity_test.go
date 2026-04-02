package domain_test

import (
	"testing"
	"time"

	domain "gct/internal/errorcode/domain"

	"github.com/google/uuid"
)

func TestNewErrorCode(t *testing.T) {
	ec := domain.NewErrorCode(
		"AUTH_FAILED", "Authentication failed", 401,
		"AUTH", "HIGH", false, 0, "Check your credentials",
	)

	if ec.Code() != "AUTH_FAILED" {
		t.Fatalf("expected code AUTH_FAILED, got %s", ec.Code())
	}
	if ec.Message() != "Authentication failed" {
		t.Fatalf("expected message Authentication failed, got %s", ec.Message())
	}
	if ec.HTTPStatus() != 401 {
		t.Fatalf("expected HTTP status 401, got %d", ec.HTTPStatus())
	}
	if ec.Category() != "AUTH" {
		t.Fatalf("expected category AUTH, got %s", ec.Category())
	}
	if ec.Severity() != "HIGH" {
		t.Fatalf("expected severity HIGH, got %s", ec.Severity())
	}
	if ec.Retryable() {
		t.Fatal("should not be retryable")
	}
	if ec.RetryAfter() != 0 {
		t.Fatalf("expected retry_after 0, got %d", ec.RetryAfter())
	}
	if ec.Suggestion() != "Check your credentials" {
		t.Fatalf("expected suggestion 'Check your credentials', got %s", ec.Suggestion())
	}

	events := ec.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventName() != "errorcode.created" {
		t.Fatalf("expected errorcode.created, got %s", events[0].EventName())
	}
}

func TestErrorCode_Update(t *testing.T) {
	ec := domain.NewErrorCode(
		"RATE_LIMIT", "Rate limit exceeded", 429,
		"SYSTEM", "MEDIUM", true, 60, "Wait and retry",
	)

	ec.Update("Too many requests", 429, "SYSTEM", "HIGH", true, 120, "Please wait 2 minutes")

	if ec.Message() != "Too many requests" {
		t.Fatalf("expected updated message, got %s", ec.Message())
	}
	if ec.Severity() != "HIGH" {
		t.Fatalf("expected updated severity HIGH, got %s", ec.Severity())
	}
	if ec.RetryAfter() != 120 {
		t.Fatalf("expected retry_after 120, got %d", ec.RetryAfter())
	}

	events := ec.Events()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
}

func TestReconstructErrorCode(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	ec := domain.ReconstructErrorCode(
		id, now, now,
		"NOT_FOUND", "Resource not found", 404,
		"DATA", "LOW", false, 0, "Check the ID",
	)

	if ec.ID() != id {
		t.Fatal("ID mismatch")
	}
	if ec.Code() != "NOT_FOUND" {
		t.Fatal("code mismatch")
	}
	if ec.HTTPStatus() != 404 {
		t.Fatal("HTTP status mismatch")
	}
	if len(ec.Events()) != 0 {
		t.Fatalf("expected 0 events, got %d", len(ec.Events()))
	}
}
