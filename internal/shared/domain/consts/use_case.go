package consts

// Structured log message constants used by use-case handlers for consistent observability.
// Each operation has a corresponding "started"/"success"/"failed" lifecycle logged at the application layer.
const (
	// Common log messages
	LogStarted = "started"
	LogSuccess = "success"
	LogFailed  = "failed"

	// User operations
	LogUserSignIn        = "user sign in"
	LogUserSignUp        = "user sign up"
	LogUserSignOut       = "user sign out"
	LogUserCreate        = "user create"
	LogUserUpdate        = "user update"
	LogUserDelete        = "user delete"
	LogUserGet           = "user get"
	LogUserGets          = "user gets"
	LogUserActivate      = "user activate"
	LogUserRotateSession = "user rotate session"
	LogUserGetByPhone    = "user get by phone"

	// Session operations
	LogSessionCreate         = "session create"
	LogSessionUpdate         = "session update"
	LogSessionDelete         = "session delete"
	LogSessionGet            = "session get"
	LogSessionGets           = "session gets"
	LogSessionRevoke         = "session revoke"
	LogSessionUpdateActivity = "session update activity"

	// Permission operations
	LogPermissionCreate       = "permission create"
	LogPermissionUpdate       = "permission update"
	LogPermissionDelete       = "permission delete"
	LogPermissionGet          = "permission get"
	LogPermissionGets         = "permission gets"
	LogPermissionAssignScope  = "permission assign scope"
	LogPermissionRemoveScope  = "permission remove scope"
	LogPermissionAssignToRole = "permission assign to role"

	// Role operations
	LogRoleCreate       = "role create"
	LogRoleUpdate       = "role update"
	LogRoleDelete       = "role delete"
	LogRoleGet          = "role get"
	LogRoleGets         = "role gets"
	LogRoleAssignPolicy = "role assign policy"

	// Policy operations
	LogPolicyCreate = "policy create"
	LogPolicyUpdate = "policy update"
	LogPolicyDelete = "policy delete"
	LogPolicyGet    = "policy get"
	LogPolicyGets   = "policy gets"

	// Scope operations
	LogScopeCreate = "scope create"
	LogScopeUpdate = "scope update"
	LogScopeDelete = "scope delete"
	LogScopeGet    = "scope get"
	LogScopeGets   = "scope gets"

	// Error reasons
	LogReasonGetUser             = "get user"
	LogReasonInvalidPassword     = "invalid password"
	LogReasonNotApproved         = "not approved"
	LogReasonGenerateRefresh     = "generate refresh token"
	LogReasonGenerateAccess      = "generate access token"
	LogReasonCreateSession       = "create session"
	LogReasonSetSessionContext   = "set session context"
	LogReasonHashPassword        = "hash password"
	LogReasonValidateInput       = "validate input"
	LogReasonInvalidKey          = "invalid key"
	LogReasonDatabaseOperation   = "database operation"
	LogReasonParseToken          = "parse token"
	LogReasonRevokeSession       = "revoke session"
	LogReasonDeleteSession       = "delete session"
	LogReasonUpdateSession       = "update session"
	LogReasonGetSession          = "get session"
	LogReasonCheckPermission     = "check permission"
	LogReasonInvalidConditionKey = "invalid policy condition key"

	// Common field names for logging
	FieldUserID    = "user_id"
	FieldSessionID = "session_id"
	FieldRoleID    = "role_id"
	FieldPermID    = "perm_id"
	FieldPolicyID  = "policy_id"
	FieldScopeID   = "scope_id"
	FieldLogin     = "login"
	FieldEmail     = "email"
	FieldPhone     = "phone"
	FieldInput     = "input"
	FieldError     = "error"
	FieldCount     = "count"
	FieldTotal     = "total"
	FieldPath      = "path"
	FieldMethod    = "method"
	FieldKey       = "key"
	FieldID        = "id"

	// Common strings
	EmptyString     = ""
	AtSymbol        = "@"
	SpaceOn         = " on "
	DefaultLanguage = "uz"
)
