package acl

import (
	"testing"

	"gct/internal/context/iam/audit/domain"

	"github.com/google/uuid"
)

func TestNewUserACL(t *testing.T) {
	acl := NewUserACL()
	if acl == nil {
		t.Fatal("expected non-nil UserACL")
	}
}

func TestUserACL_ToAuditLog_ValidResourceID(t *testing.T) {
	acl := NewUserACL()

	userID := uuid.New()
	resourceID := uuid.New()

	log := acl.ToAuditLog(userID, domain.AuditActionUserCreate, "user", resourceID.String())

	if log == nil {
		t.Fatal("expected non-nil AuditLog")
	}
	if log.UserID() == nil || *log.UserID() != userID {
		t.Fatalf("expected userID %v, got %v", userID, log.UserID())
	}
	if log.Action() != domain.AuditActionUserCreate {
		t.Fatalf("expected action %v, got %v", domain.AuditActionUserCreate, log.Action())
	}
	if log.ResourceType() == nil || *log.ResourceType() != "user" {
		t.Fatalf("expected resource type 'user', got %v", log.ResourceType())
	}
	if log.ResourceID() == nil || *log.ResourceID() != resourceID {
		t.Fatalf("expected resource ID %v, got %v", resourceID, log.ResourceID())
	}
	if !log.Success() {
		t.Fatal("expected success=true")
	}
	if log.SessionID() != nil {
		t.Fatalf("expected nil sessionID, got %v", log.SessionID())
	}
	if log.Platform() != nil {
		t.Fatalf("expected nil platform, got %v", log.Platform())
	}
	if log.IPAddress() != nil {
		t.Fatalf("expected nil ipAddress, got %v", log.IPAddress())
	}
	if log.UserAgent() != nil {
		t.Fatalf("expected nil userAgent, got %v", log.UserAgent())
	}
	if log.Permission() != nil {
		t.Fatalf("expected nil permission, got %v", log.Permission())
	}
	if log.PolicyID() != nil {
		t.Fatalf("expected nil policyID, got %v", log.PolicyID())
	}
	if log.Decision() != nil {
		t.Fatalf("expected nil decision, got %v", log.Decision())
	}
	if log.ErrorMessage() != nil {
		t.Fatalf("expected nil errorMessage, got %v", log.ErrorMessage())
	}
}

func TestUserACL_ToAuditLog_InvalidResourceID(t *testing.T) {
	acl := NewUserACL()

	userID := uuid.New()

	log := acl.ToAuditLog(userID, domain.AuditActionUserUpdate, "user", "not-a-uuid")

	if log == nil {
		t.Fatal("expected non-nil AuditLog")
	}
	if log.ResourceID() != nil {
		t.Fatalf("expected nil resourceID for invalid UUID string, got %v", log.ResourceID())
	}
	if log.Action() != domain.AuditActionUserUpdate {
		t.Fatalf("expected action %v, got %v", domain.AuditActionUserUpdate, log.Action())
	}
}

func TestUserACL_ToAuditLog_EmptyResourceID(t *testing.T) {
	acl := NewUserACL()

	userID := uuid.New()

	log := acl.ToAuditLog(userID, domain.AuditActionUserDelete, "user", "")

	if log == nil {
		t.Fatal("expected non-nil AuditLog")
	}
	if log.ResourceID() != nil {
		t.Fatalf("expected nil resourceID for empty string, got %v", log.ResourceID())
	}
}

func TestUserACL_ToAuditLog_DifferentActions(t *testing.T) {
	acl := NewUserACL()
	userID := uuid.New()
	resID := uuid.New()

	actions := []domain.AuditAction{
		domain.AuditActionLogin,
		domain.AuditActionLogout,
		domain.AuditActionRoleAssign,
		domain.AuditActionPasswordChange,
	}

	for _, action := range actions {
		log := acl.ToAuditLog(userID, action, "user", resID.String())
		if log.Action() != action {
			t.Fatalf("expected action %v, got %v", action, log.Action())
		}
	}
}

func TestUserACL_ToAuditLog_MetadataNotNil(t *testing.T) {
	acl := NewUserACL()
	userID := uuid.New()

	log := acl.ToAuditLog(userID, domain.AuditActionUserCreate, "user", uuid.New().String())

	if log.Metadata() == nil {
		t.Fatal("expected non-nil metadata map")
	}
}
