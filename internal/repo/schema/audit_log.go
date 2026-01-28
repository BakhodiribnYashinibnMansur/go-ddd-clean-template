package schema

// Table name
const TableAuditLog = "audit_log"

// AuditLog table columns
const (
	AuditLogID           = "id"
	AuditLogUserID       = "user_id"
	AuditLogSessionID    = "session_id"
	AuditLogAction       = "action"
	AuditLogResourceType = "resource_type"
	AuditLogResourceID   = "resource_id"
	AuditLogPlatform     = "platform"
	AuditLogIPAddress    = "ip_address"
	AuditLogUserAgent    = "user_agent"
	AuditLogPermission   = "permission"
	AuditLogPolicyID     = "policy_id"
	AuditLogDecision     = "decision"
	AuditLogSuccess      = "success"
	AuditLogErrorMessage = "error_message"
	AuditLogMetadata     = "metadata"
	AuditLogCreatedAt    = "created_at"
)
