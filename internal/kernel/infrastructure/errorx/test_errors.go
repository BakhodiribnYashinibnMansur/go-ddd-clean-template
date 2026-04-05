// Package errors provides static test errors for use in test files.
// This ensures compliance with the err113 linter rule which requires
// wrapped static errors instead of dynamic errors.
package errorx

import "errors"

// Common test errors that can be wrapped with context in test files.
var (
	// Database errors
	ErrTestDB                = errors.New("database error")
	ErrTestDBConnection      = errors.New("database connection error")
	ErrTestConnectionTimeout = errors.New("connection timeout")
	ErrTestSelectQuery       = errors.New("select query error")
	ErrTestCountError        = errors.New("count error")

	// Constraint errors
	ErrTestDuplicateKey     = errors.New("duplicate key value violates unique constraint")
	ErrTestForeignKey       = errors.New("foreign key constraint violation")
	ErrTestUniqueConstraint = errors.New("unique constraint violation")
	ErrTestCheckConstraint  = errors.New("check constraint violation")

	// Validation errors
	ErrTestInvalidPhone     = errors.New("invalid phone format")
	ErrTestInvalidUserID    = errors.New("invalid user ID")
	ErrTestInvalidCompanyID = errors.New("invalid company ID")
	ErrTestInvalidFilter    = errors.New("invalid filter")
	ErrTestInvalidSession   = errors.New("invalid session")
	ErrTestValidationFailed = errors.New("validation failed")

	// Password errors
	ErrTestPasswordHashEmpty = errors.New("password hash cannot be empty")

	// Permission errors
	ErrTestPermissionDenied = errors.New("permission denied for table")

	// Lock errors
	ErrTestDatabaseLocked = errors.New("database is locked")

	// Redis errors
	ErrTestRedisConnection      = errors.New("redis connection failed")
	ErrTestRedisScan            = errors.New("redis scan error")
	ErrTestRedisDialTCP         = errors.New("dial tcp: connection refused")
	ErrTestRedisIOTimeout       = errors.New("i/o timeout")
	ErrTestRedisContextDeadline = errors.New("context deadline exceeded")
	ErrTestRedisWrongPass       = errors.New("WRONGPASS invalid username-password pair")
	ErrTestRedisWrongType       = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	ErrTestRedisOOM             = errors.New("OOM command not allowed when used memory > 'maxmemory'")
	ErrTestRedisReadOnly        = errors.New("READONLY You can't write against a read only replica")
	ErrTestRedisClusterDown     = errors.New("CLUSTERDOWN The cluster is down")
	ErrTestRedisNoScript        = errors.New("NOSCRIPT No matching script. Please use EVAL")
	ErrTestRedisNoAuth          = errors.New("NOAUTH Authentication required")
	ErrTestRedisEOF             = errors.New("EOF")
	ErrTestRedisBrokenPipe      = errors.New("write: broken pipe")
	ErrTestRedisRandom          = errors.New("some random redis error")
	ErrTestRedisError           = errors.New("some redis error")

	// User/Session errors
	ErrTestUserNotFound          = errors.New("user not found")
	ErrTestSessionNotFound       = errors.New("session not found")
	ErrTestSessionCreationFailed = errors.New("session creation failed")
	ErrTestSessionRevokeFailed   = errors.New("session revoke failed")

	// Operation errors
	ErrTestDeleteFailed = errors.New("delete failed")
	ErrTestUpdateFailed = errors.New("update failed")
	ErrTestScanError    = errors.New("scan error")

	// Generic errors
	ErrTestHandler = errors.New("handler error")
	ErrTestService = errors.New("service error")
	ErrTestRegular = errors.New("regular error")
	ErrTestBase    = errors.New("base error")
)
