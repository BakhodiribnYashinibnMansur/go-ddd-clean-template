package domain_test

import (
	"testing"
	"time"

	"gct/internal/context/iam/supporting/audit/domain"

	"github.com/google/uuid"
)

func TestNewEndpointHistory(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	ip := "192.168.1.1"
	ua := "Mozilla/5.0"

	h := domain.NewEndpointHistory(
		&userID, "/api/users", "GET", 200, 42, &ip, &ua,
	)

	if h.ID() == uuid.Nil {
		t.Fatal("expected non-nil ID")
	}
	if h.UserID() == nil || *h.UserID() != userID {
		t.Fatal("userID mismatch")
	}
	if h.Endpoint() != "/api/users" {
		t.Errorf("expected endpoint '/api/users', got %q", h.Endpoint())
	}
	if h.Method() != "GET" {
		t.Errorf("expected method 'GET', got %q", h.Method())
	}
	if h.StatusCode() != 200 {
		t.Errorf("expected status 200, got %d", h.StatusCode())
	}
	if h.Latency() != 42 {
		t.Errorf("expected latency 42, got %d", h.Latency())
	}
	if h.IPAddress() == nil || *h.IPAddress() != "192.168.1.1" {
		t.Error("ipAddress mismatch")
	}
	if h.UserAgent() == nil || *h.UserAgent() != "Mozilla/5.0" {
		t.Error("userAgent mismatch")
	}
	if h.GetCreatedAt().IsZero() {
		t.Error("expected non-zero createdAt")
	}
}

func TestNewEndpointHistory_NilOptionals(t *testing.T) {
	t.Parallel()

	h := domain.NewEndpointHistory(
		nil, "/health", "GET", 200, 1, nil, nil,
	)

	if h.UserID() != nil {
		t.Error("expected nil userID")
	}
	if h.IPAddress() != nil {
		t.Error("expected nil ipAddress")
	}
	if h.UserAgent() != nil {
		t.Error("expected nil userAgent")
	}
}

func TestReconstructEndpointHistory(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	userID := uuid.New()
	now := time.Now()
	ip := "10.0.0.1"
	ua := "curl/7.68"

	h := domain.ReconstructEndpointHistory(
		id, now,
		&userID, "/api/roles", "POST", 201, 150, &ip, &ua,
	)

	if h.ID() != id {
		t.Fatalf("expected ID %s, got %s", id, h.ID())
	}
	if h.GetCreatedAt() != now {
		t.Error("expected createdAt to match")
	}
	if *h.UserID() != userID {
		t.Error("userID mismatch")
	}
	if h.Endpoint() != "/api/roles" {
		t.Errorf("expected endpoint '/api/roles', got %q", h.Endpoint())
	}
	if h.Method() != "POST" {
		t.Errorf("expected method 'POST', got %q", h.Method())
	}
	if h.StatusCode() != 201 {
		t.Errorf("expected status 201, got %d", h.StatusCode())
	}
	if h.Latency() != 150 {
		t.Errorf("expected latency 150, got %d", h.Latency())
	}
	if *h.IPAddress() != "10.0.0.1" {
		t.Error("ipAddress mismatch")
	}
	if *h.UserAgent() != "curl/7.68" {
		t.Error("userAgent mismatch")
	}
}

func TestReconstructEndpointHistory_NilOptionals(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	now := time.Now()

	h := domain.ReconstructEndpointHistory(
		id, now,
		nil, "/health", "GET", 200, 5, nil, nil,
	)

	if h.UserID() != nil {
		t.Error("expected nil userID")
	}
	if h.IPAddress() != nil {
		t.Error("expected nil ipAddress")
	}
	if h.UserAgent() != nil {
		t.Error("expected nil userAgent")
	}
}

func TestNewEndpointHistory_ErrorStatus(t *testing.T) {
	t.Parallel()

	h := domain.NewEndpointHistory(
		nil, "/api/fail", "DELETE", 500, 2000, nil, nil,
	)

	if h.StatusCode() != 500 {
		t.Errorf("expected status 500, got %d", h.StatusCode())
	}
	if h.Latency() != 2000 {
		t.Errorf("expected latency 2000, got %d", h.Latency())
	}
}
