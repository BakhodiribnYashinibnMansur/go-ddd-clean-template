package acl

import (
	"gct/internal/context/iam/supporting/audit/domain"

	"github.com/google/uuid"
)

// UserACL translates User-bounded-context events into Audit domain objects.
type UserACL struct{}

// NewUserACL creates a new UserACL.
func NewUserACL() *UserACL {
	return &UserACL{}
}

// ToAuditLog creates an AuditLog from user-related event data.
func (a *UserACL) ToAuditLog(
	userID uuid.UUID,
	action domain.AuditAction,
	resourceType string,
	resourceID string,
) *domain.AuditLog {
	resType := resourceType
	resID, err := uuid.Parse(resourceID)

	var resIDPtr *uuid.UUID
	if err == nil {
		resIDPtr = &resID
	}

	return domain.NewAuditLog(
		&userID,
		nil,    // sessionID
		action,
		&resType,
		resIDPtr,
		nil, // platform
		nil, // ipAddress
		nil, // userAgent
		nil, // permission
		nil, // policyID
		nil, // decision
		true,
		nil, // errorMessage
		nil, // metadata
	)
}
