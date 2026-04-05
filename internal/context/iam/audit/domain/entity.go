package domain

import (
	"time"

	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
)

// AuditAction is a closed set of auditable business actions stored as PostgreSQL-compatible string constants.
// New actions require adding a constant here and ensuring any persistence layer accepts the new value.
type AuditAction string

const (
	AuditActionLogin           AuditAction = "LOGIN"
	AuditActionLogout          AuditAction = "LOGOUT"
	AuditActionSessionRevoke   AuditAction = "SESSION_REVOKE"
	AuditActionPasswordChange  AuditAction = "PASSWORD_CHANGE"
	AuditActionMfaVerifyFail   AuditAction = "MFA_VERIFY_FAIL"
	AuditActionAccessGranted   AuditAction = "ACCESS_GRANTED"
	AuditActionAccessDenied    AuditAction = "ACCESS_DENIED"
	AuditActionPolicyMatched   AuditAction = "POLICY_MATCHED"
	AuditActionPolicyDenied    AuditAction = "POLICY_DENIED"
	AuditActionUserCreate      AuditAction = "USER_CREATE"
	AuditActionUserUpdate      AuditAction = "USER_UPDATE"
	AuditActionUserDelete      AuditAction = "USER_DELETE"
	AuditActionRoleAssign      AuditAction = "ROLE_ASSIGN"
	AuditActionRoleRemove      AuditAction = "ROLE_REMOVE"
	AuditActionOrderApprove    AuditAction = "ORDER_APPROVE"
	AuditActionOrderCancel     AuditAction = "ORDER_CANCEL"
	AuditActionPaymentProcess  AuditAction = "PAYMENT_PROCESS"
	AuditActionPaymentCancel   AuditAction = "PAYMENT_CANCEL"
	AuditActionPolicyEvaluated AuditAction = "POLICY_EVALUATED"
	AuditActionAdminChange     AuditAction = "ADMIN_CHANGE"
)

// AuditLog is the aggregate root for audit log entries.
// It is intentionally immutable after creation — there are no Update or Delete methods.
// The metadata map captures arbitrary contextual data (request bodies, policy details) that
// varies per action type, keeping the schema flexible without requiring column changes.
type AuditLog struct {
	shared.AggregateRoot
	userID       *uuid.UUID
	sessionID    *uuid.UUID
	action       AuditAction
	resourceType *string
	resourceID   *uuid.UUID
	platform     *string
	ipAddress    *string
	userAgent    *string
	permission   *string
	policyID     *uuid.UUID
	decision     *string
	success      bool
	errorMessage *string
	metadata     map[string]string
}

// NewAuditLog creates a new AuditLog aggregate.
func NewAuditLog(
	userID *uuid.UUID,
	sessionID *uuid.UUID,
	action AuditAction,
	resourceType *string,
	resourceID *uuid.UUID,
	platform *string,
	ipAddress *string,
	userAgent *string,
	permission *string,
	policyID *uuid.UUID,
	decision *string,
	success bool,
	errorMessage *string,
	metadata map[string]string,
) *AuditLog {
	if metadata == nil {
		metadata = make(map[string]string)
	}

	a := &AuditLog{
		AggregateRoot: shared.NewAggregateRoot(),
		userID:        userID,
		sessionID:     sessionID,
		action:        action,
		resourceType:  resourceType,
		resourceID:    resourceID,
		platform:      platform,
		ipAddress:     ipAddress,
		userAgent:     userAgent,
		permission:    permission,
		policyID:      policyID,
		decision:      decision,
		success:       success,
		errorMessage:  errorMessage,
		metadata:      metadata,
	}

	a.AddEvent(NewAuditLogCreated(a.ID(), action))

	return a
}

// ReconstructAuditLog rebuilds an AuditLog from persisted data.
func ReconstructAuditLog(
	id uuid.UUID,
	createdAt time.Time,
	userID *uuid.UUID,
	sessionID *uuid.UUID,
	action AuditAction,
	resourceType *string,
	resourceID *uuid.UUID,
	platform *string,
	ipAddress *string,
	userAgent *string,
	permission *string,
	policyID *uuid.UUID,
	decision *string,
	success bool,
	errorMessage *string,
	metadata map[string]string,
) *AuditLog {
	if metadata == nil {
		metadata = make(map[string]string)
	}

	return &AuditLog{
		AggregateRoot: shared.NewAggregateRootWithID(id, createdAt, createdAt, nil),
		userID:        userID,
		sessionID:     sessionID,
		action:        action,
		resourceType:  resourceType,
		resourceID:    resourceID,
		platform:      platform,
		ipAddress:     ipAddress,
		userAgent:     userAgent,
		permission:    permission,
		policyID:      policyID,
		decision:      decision,
		success:       success,
		errorMessage:  errorMessage,
		metadata:      metadata,
	}
}

// Getters

func (a *AuditLog) UserID() *uuid.UUID        { return a.userID }
func (a *AuditLog) SessionID() *uuid.UUID      { return a.sessionID }
func (a *AuditLog) Action() AuditAction        { return a.action }
func (a *AuditLog) ResourceType() *string       { return a.resourceType }
func (a *AuditLog) ResourceID() *uuid.UUID      { return a.resourceID }
func (a *AuditLog) Platform() *string            { return a.platform }
func (a *AuditLog) IPAddress() *string           { return a.ipAddress }
func (a *AuditLog) UserAgent() *string           { return a.userAgent }
func (a *AuditLog) Permission() *string          { return a.permission }
func (a *AuditLog) PolicyID() *uuid.UUID         { return a.policyID }
func (a *AuditLog) Decision() *string            { return a.decision }
func (a *AuditLog) Success() bool                { return a.success }
func (a *AuditLog) ErrorMessage() *string        { return a.errorMessage }
func (a *AuditLog) Metadata() map[string]string   { return a.metadata }
