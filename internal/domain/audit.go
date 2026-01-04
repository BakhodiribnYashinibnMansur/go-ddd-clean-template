package domain

import (
	"time"

	"github.com/google/uuid"
)

type AuditActionType string

const (
	AuditActionLogin           AuditActionType = "LOGIN"
	AuditActionLogout          AuditActionType = "LOGOUT"
	AuditActionSessionRevoke   AuditActionType = "SESSION_REVOKE"
	AuditActionPasswordChange  AuditActionType = "PASSWORD_CHANGE"
	AuditActionMfaVerifyFail   AuditActionType = "MFA_VERIFY_FAIL"
	AuditActionAccessGranted   AuditActionType = "ACCESS_GRANTED"
	AuditActionAccessDenied    AuditActionType = "ACCESS_DENIED"
	AuditActionPolicyMatched   AuditActionType = "POLICY_MATCHED"
	AuditActionPolicyDenied    AuditActionType = "POLICY_DENIED"
	AuditActionUserCreate      AuditActionType = "USER_CREATE"
	AuditActionUserUpdate      AuditActionType = "USER_UPDATE"
	AuditActionUserDelete      AuditActionType = "USER_DELETE"
	AuditActionRoleAssign      AuditActionType = "ROLE_ASSIGN"
	AuditActionRoleRemove      AuditActionType = "ROLE_REMOVE"
	AuditActionOrderApprove    AuditActionType = "ORDER_APPROVE"
	AuditActionOrderCancel     AuditActionType = "ORDER_CANCEL"
	AuditActionPaymentProcess  AuditActionType = "PAYMENT_PROCESS"
	AuditActionPaymentCancel   AuditActionType = "PAYMENT_CANCEL"
	AuditActionPolicyEvaluated AuditActionType = "POLICY_EVALUATED"
)

type AuditLog struct {
	ID           uuid.UUID       `db:"id"            json:"id"`
	UserID       *uuid.UUID      `db:"user_id"       json:"user_id,omitempty"`
	SessionID    *uuid.UUID      `db:"session_id"    json:"session_id,omitempty"`
	Action       AuditActionType `db:"action"        json:"action"`
	ResourceType *string         `db:"resource_type" json:"resource_type,omitempty"`
	ResourceID   *uuid.UUID      `db:"resource_id"   json:"resource_id,omitempty"`
	Platform     *string         `db:"platform"      json:"platform,omitempty"`
	IPAddress    *string         `db:"ip_address"    json:"ip_address,omitempty"`
	UserAgent    *string         `db:"user_agent"    json:"user_agent,omitempty"`
	Permission   *string         `db:"permission"    json:"permission,omitempty"`
	PolicyID     *uuid.UUID      `db:"policy_id"     json:"policy_id,omitempty"`
	Decision     *string         `db:"decision"      json:"decision,omitempty"`
	Success      bool            `db:"success"       json:"success"`
	ErrorMessage *string         `db:"error_message" json:"error_message,omitempty"`
	Metadata     map[string]any  `db:"metadata"      json:"metadata,omitempty"`
	CreatedAt    time.Time       `db:"created_at"    json:"created_at"`
}

type AuditLogFilter struct {
	UserID       *uuid.UUID       `json:"user_id,omitempty"`
	Action       *AuditActionType `json:"action,omitempty"`
	ResourceType *string          `json:"resource_type,omitempty"`
	ResourceID   *uuid.UUID       `json:"resource_id,omitempty"`
	Success      *bool            `json:"success,omitempty"`
	FromDate     *time.Time       `json:"from_date,omitempty"`
	ToDate       *time.Time       `json:"to_date,omitempty"`
}

type AuditLogsFilter struct {
	AuditLogFilter
	Pagination *Pagination `json:"pagination"`
}

type EndpointHistory struct {
	ID           uuid.UUID  `db:"id"            json:"id"`
	UserID       *uuid.UUID `db:"user_id"       json:"user_id,omitempty"`
	SessionID    *uuid.UUID `db:"session_id"    json:"session_id,omitempty"`
	Method       string     `db:"method"        json:"method"`
	Path         string     `db:"path"          json:"path"`
	StatusCode   int        `db:"status_code"   json:"status_code"`
	DurationMs   int        `db:"duration_ms"   json:"duration_ms"`
	Platform     *string    `db:"platform"      json:"platform,omitempty"`
	IPAddress    *string    `db:"ip_address"    json:"ip_address,omitempty"`
	UserAgent    *string    `db:"user_agent"    json:"user_agent,omitempty"`
	Permission   *string    `db:"permission"    json:"permission,omitempty"`
	Decision     *string    `db:"decision"      json:"decision,omitempty"`
	RequestID    *uuid.UUID `db:"request_id"    json:"request_id,omitempty"`
	RateLimited  bool       `db:"rate_limited"  json:"rate_limited"`
	ResponseSize *int       `db:"response_size" json:"response_size,omitempty"`
	ErrorMessage *string    `db:"error_message" json:"error_message,omitempty"`
	CreatedAt    time.Time  `db:"created_at"    json:"created_at"`
}

type EndpointHistoryFilter struct {
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	Method     *string    `json:"method,omitempty"`
	Path       *string    `json:"path,omitempty"`
	StatusCode *int       `json:"status_code,omitempty"`
	FromDate   *time.Time `json:"from_date,omitempty"`
	ToDate     *time.Time `json:"to_date,omitempty"`
}

type EndpointHistoriesFilter struct {
	EndpointHistoryFilter
	Pagination *Pagination `json:"pagination"`
}
