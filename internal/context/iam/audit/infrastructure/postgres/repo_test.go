package postgres

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewAuditLogWriteRepo(t *testing.T) {
	repo := NewAuditLogWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil AuditLogWriteRepo")
	}
}

func TestNewEndpointHistoryWriteRepo(t *testing.T) {
	repo := NewEndpointHistoryWriteRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil EndpointHistoryWriteRepo")
	}
}

func TestNewAuditReadRepo(t *testing.T) {
	repo := NewAuditReadRepo(nil)
	if repo == nil {
		t.Fatal("expected non-nil AuditReadRepo")
	}
}
