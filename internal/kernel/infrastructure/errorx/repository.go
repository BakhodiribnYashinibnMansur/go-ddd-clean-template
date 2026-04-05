package errorx

// ============================================================================
// Repository Layer Error Codes
// ============================================================================

const (
	// Database errors
	ErrRepoNotFound  = "REPO_NOT_FOUND"
	CodeRepoNotFound = "2001"

	ErrRepoAlreadyExists  = "REPO_ALREADY_EXISTS"
	CodeRepoAlreadyExists = "2002"

	ErrRepoDatabase  = "REPO_DATABASE_ERROR"
	CodeRepoDatabase = "2003"

	ErrRepoTimeout  = "REPO_TIMEOUT"
	CodeRepoTimeout = "2004"

	ErrRepoConnection  = "REPO_CONNECTION_ERROR"
	CodeRepoConnection = "2005"

	ErrRepoTransaction  = "REPO_TRANSACTION_ERROR"
	CodeRepoTransaction = "2006"

	ErrRepoConstraint  = "REPO_CONSTRAINT_VIOLATION"
	CodeRepoConstraint = "2007"

	ErrRepoUnknown  = "REPO_UNKNOWN_ERROR"
	CodeRepoUnknown = "2099"
)

// Repository error messages
var repoMessages = map[string]string{
	ErrRepoNotFound:      "Record not found in database",
	ErrRepoAlreadyExists: "Record already exists",
	ErrRepoDatabase:      "Database error occurred",
	ErrRepoTimeout:       "Database operation timeout",
	ErrRepoConnection:    "Database connection error",
	ErrRepoTransaction:   "Transaction error",
	ErrRepoConstraint:    "Database constraint violation",
	ErrRepoUnknown:       "Unknown repository error",
}

// NewRepoError creates a new repository error
func NewRepoError(code, message string) *AppError {
	return New(code, message)
}

// WrapRepoError wraps an error as repository error
func WrapRepoError(err error, code, message string) *AppError {
	return Wrap(err, code, message)
}
