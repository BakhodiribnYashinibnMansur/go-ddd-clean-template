package middleware

import (
	"testing"
)

// Idempotency middleware requires a real *redis.Client and performs
// Redis operations (Get, SetNX, Set, Del) that cannot be easily mocked
// without introducing an interface wrapper around the Redis client.
//
// Skipping unit tests for this middleware. Integration tests with a real
// Redis instance (e.g. via testcontainers) are recommended.

func TestIdempotency_Constants(t *testing.T) {
	if HeaderIdempotencyKey != "Idempotency-Key" {
		t.Errorf("unexpected HeaderIdempotencyKey: %s", HeaderIdempotencyKey)
	}
	if IdempotencyKeyPrefix != "idempotency:" {
		t.Errorf("unexpected IdempotencyKeyPrefix: %s", IdempotencyKeyPrefix)
	}
}

func TestResponseWriter_Interface(t *testing.T) {
	// Verify the responseWriter struct has the expected body field
	rw := &responseWriter{}
	if rw.body != nil {
		t.Error("expected nil body for zero-value responseWriter")
	}
}
