package domain_test

import (
	"testing"
	"time"

	"gct/internal/context/iam/audit/domain"

	"github.com/google/uuid"
)

func TestNewAuditLog_Defaults(t *testing.T) {
	userID := uuid.New()
	action := domain.AuditActionLogin

	a := domain.NewAuditLog(
		&userID, nil, action,
		nil, nil, nil, nil, nil, nil, nil, nil,
		true, nil, nil,
	)

	if a.UserID() == nil || *a.UserID() != userID {
		t.Fatal("userID mismatch")
	}
	if a.SessionID() != nil {
		t.Fatal("sessionID should be nil")
	}
	if a.Action() != domain.AuditActionLogin {
		t.Fatalf("expected action LOGIN, got %s", a.Action())
	}
	if !a.Success() {
		t.Fatal("expected success to be true")
	}
	if a.ResourceType() != nil {
		t.Fatal("resourceType should be nil")
	}
	if a.ResourceID() != nil {
		t.Fatal("resourceID should be nil")
	}
	if a.Platform() != nil {
		t.Fatal("platform should be nil")
	}
	if a.IPAddress() != nil {
		t.Fatal("ipAddress should be nil")
	}
	if a.UserAgent() != nil {
		t.Fatal("userAgent should be nil")
	}
	if a.Permission() != nil {
		t.Fatal("permission should be nil")
	}
	if a.PolicyID() != nil {
		t.Fatal("policyID should be nil")
	}
	if a.Decision() != nil {
		t.Fatal("decision should be nil")
	}
	if a.ErrorMessage() != nil {
		t.Fatal("errorMessage should be nil")
	}
	if a.Metadata() == nil {
		t.Fatal("metadata should not be nil")
	}
	if len(a.Metadata()) != 0 {
		t.Fatal("metadata should be empty")
	}
	if a.ID() == uuid.Nil {
		t.Fatal("ID should be generated")
	}
}

func TestNewAuditLog_AllFields(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	resourceID := uuid.New()
	policyID := uuid.New()
	resType := "user"
	platform := "web"
	ip := "192.168.1.1"
	ua := "Mozilla/5.0"
	perm := "users:read"
	decision := "allow"
	errMsg := "some error"
	meta := map[string]string{"key": "value"}

	a := domain.NewAuditLog(
		&userID, &sessionID, domain.AuditActionUserCreate,
		&resType, &resourceID, &platform, &ip, &ua, &perm, &policyID, &decision,
		false, &errMsg, meta,
	)

	if *a.UserID() != userID {
		t.Fatal("userID mismatch")
	}
	if *a.SessionID() != sessionID {
		t.Fatal("sessionID mismatch")
	}
	if a.Action() != domain.AuditActionUserCreate {
		t.Fatalf("expected action USER_CREATE, got %s", a.Action())
	}
	if *a.ResourceType() != "user" {
		t.Fatal("resourceType mismatch")
	}
	if *a.ResourceID() != resourceID {
		t.Fatal("resourceID mismatch")
	}
	if *a.Platform() != "web" {
		t.Fatal("platform mismatch")
	}
	if *a.IPAddress() != "192.168.1.1" {
		t.Fatal("ipAddress mismatch")
	}
	if *a.UserAgent() != "Mozilla/5.0" {
		t.Fatal("userAgent mismatch")
	}
	if *a.Permission() != "users:read" {
		t.Fatal("permission mismatch")
	}
	if *a.PolicyID() != policyID {
		t.Fatal("policyID mismatch")
	}
	if *a.Decision() != "allow" {
		t.Fatal("decision mismatch")
	}
	if a.Success() {
		t.Fatal("expected success to be false")
	}
	if *a.ErrorMessage() != "some error" {
		t.Fatal("errorMessage mismatch")
	}
	if a.Metadata()["key"] != "value" {
		t.Fatal("metadata mismatch")
	}
}

func TestNewAuditLog_EventPublished(t *testing.T) {
	a := domain.NewAuditLog(
		nil, nil, domain.AuditActionLogout,
		nil, nil, nil, nil, nil, nil, nil, nil,
		true, nil, nil,
	)

	events := a.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].EventName() != "audit_log.created" {
		t.Fatalf("expected event audit_log.created, got %s", events[0].EventName())
	}

	if events[0].AggregateID() != a.ID() {
		t.Fatal("event aggregate ID should match audit log ID")
	}
}

func TestReconstructAuditLog(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	now := time.Now()

	a := domain.ReconstructAuditLog(
		id, now,
		&userID, nil, domain.AuditActionPasswordChange,
		nil, nil, nil, nil, nil, nil, nil, nil,
		true, nil, nil,
	)

	if a.ID() != id {
		t.Fatal("ID mismatch")
	}
	if *a.UserID() != userID {
		t.Fatal("userID mismatch")
	}
	if a.Action() != domain.AuditActionPasswordChange {
		t.Fatal("action mismatch")
	}

	// Reconstructed audit logs should have no pending events.
	if len(a.Events()) != 0 {
		t.Fatal("reconstructed audit log should have no events")
	}
}
