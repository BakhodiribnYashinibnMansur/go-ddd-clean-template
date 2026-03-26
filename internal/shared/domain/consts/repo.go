package consts

// Repository-layer error messages and SQL column name constants.
// Error messages are used in fmt.Errorf wrapping to provide consistent context in logs.
// Column names prevent typos in query builder calls across all repository implementations.
const (
	// Common error messages
	ErrMsgFailedToBuildQuery       = "failed to build select query"
	ErrMsgFailedToBuildInsert      = "failed to build insert query"
	ErrMsgFailedToBuildUpdate      = "failed to build update query"
	ErrMsgFailedToBuildDelete      = "failed to build delete query"
	ErrMsgFailedToExecuteQuery     = "failed to execute query"
	ErrMsgFailedToScanRow          = "failed to scan row"
	ErrMsgFailedToBeginTx          = "failed to begin transaction"
	ErrMsgFailedToCommitTx         = "failed to commit transaction"
	ErrMsgFailedToRollbackTx       = "failed to rollback transaction"
	ErrMsgFailedToMarshalJSON      = "failed to marshal JSON"
	ErrMsgFailedToUnmarshalJSON    = "failed to unmarshal JSON"
	ErrMsgInvalidFilter            = "invalid filter"
	ErrMsgInvalidInput             = "invalid input"
	ErrMsgRecordNotFound           = "record not found"
	ErrMsgDuplicateKey             = "duplicate key"
	ErrMsgForeignKeyViolation      = "foreign key violation"
	ErrMsgCheckConstraintViolation = "check constraint violation"

	// Common SQL column names
	ColID           = "id"
	ColCreatedAt    = "created_at"
	ColUpdatedAt    = "updated_at"
	ColDeletedAt    = "deleted_at"
	ColPath         = "path"
	ColMethod       = "method"
	ColName         = "name"
	ColDescription  = "description"
	ColUserID       = "user_id"
	ColSessionID    = "session_id"
	ColRoleID       = "role_id"
	ColPermissionID = "permission_id"
	ColPolicyID     = "policy_id"
	ColScopeID      = "scope_id"

	// Pagination
	ColLimit  = "limit"
	ColOffset = "offset"

	// SQL-specific operators (different from policy operators)
	SQLOpLike    = "LIKE"
	SQLOrderAsc  = "ASC"
	SQLOrderDesc = "DESC"
)
