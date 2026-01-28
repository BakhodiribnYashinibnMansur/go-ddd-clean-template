package errors

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
)

// HandleMySQLError handles MySQL errors and converts them to AppError
// This centralizes all MySQL error handling logic
func HandleMySQLError(err error, table string, extraFields map[string]any) *AppError {
	if err == nil {
		return nil
	}

	// ============================================================================
	// Special Case: sql.ErrNoRows (Record not found)
	// ============================================================================
	if errors.Is(err, sql.ErrNoRows) {
		appErr := AutoSource(
			NewRepoError(ErrRepoNotFound,
				"record not found"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithDetails("No matching record found in database")

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr
	}

	// Try to cast to mysql.MySQLError
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return handleMySQLSpecificError(mysqlErr, table, extraFields)
	}

	// Check for connection errors in error message
	errMsg := err.Error()
	if isMySQLConnectionError(errMsg) {
		return handleMySQLConnectionError(table, extraFields)
	}

	// ============================================================================
	// Default: Generic MySQL error
	// ============================================================================
	appErr := AutoSource(
		WrapRepoError(err, ErrRepoDatabase,
			"mysql operation failed"))

	if table != "" {
		_ = appErr.WithField("table", table)
	}
	_ = appErr.WithDetails(errMsg)

	for key, value := range extraFields {
		_ = appErr.WithField(key, value)
	}
	return appErr
}

// handleMySQLSpecificError handles MySQL-specific errors by error code
func handleMySQLSpecificError(mysqlErr *mysql.MySQLError, table string, extraFields map[string]any) *AppError {
	switch mysqlErr.Number {
	// ============================================================================
	// 1062 — Duplicate Entry (Unique Violation)
	// ============================================================================
	case 1062:
		appErr := AutoSource(
			NewRepoError(ErrRepoAlreadyExists,
				"duplicate entry"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("mysql_code", mysqlErr.Number).
			WithField("sql_state", mysqlErr.SQLState[0]).
			WithDetails("A record with this unique constraint already exists")

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr

	// ============================================================================
	// 1452 — Foreign Key Constraint Fails
	// ============================================================================
	case 1452:
		appErr := AutoSource(
			NewRepoError(ErrRepoConstraint,
				"foreign key constraint violation"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("mysql_code", mysqlErr.Number).
			WithField("sql_state", mysqlErr.SQLState[0]).
			WithDetails(mysqlErr.Message)

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr

	// ============================================================================
	// 1451 — Cannot Delete or Update Parent Row (Foreign Key Violation)
	// ============================================================================
	case 1451:
		appErr := AutoSource(
			NewRepoError(ErrRepoConstraint,
				"cannot delete or update parent row"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("mysql_code", mysqlErr.Number).
			WithField("sql_state", mysqlErr.SQLState[0]).
			WithDetails("Foreign key constraint prevents this operation")

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr

	// ============================================================================
	// 1048 — Column Cannot Be NULL
	// ============================================================================
	case 1048:
		appErr := AutoSource(
			NewRepoError(ErrRepoConstraint,
				"null value constraint violation"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("mysql_code", mysqlErr.Number).
			WithField("sql_state", mysqlErr.SQLState[0]).
			WithDetails("Required field cannot be null")

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr

	// ============================================================================
	// 1146 — Table Doesn't Exist
	// ============================================================================
	case 1146:
		appErr := AutoSource(
			NewRepoError(ErrRepoDatabase,
				"table does not exist"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("mysql_code", mysqlErr.Number).
			WithField("sql_state", mysqlErr.SQLState[0]).
			WithDetails(mysqlErr.Message)

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr

	// ============================================================================
	// 1054 — Unknown Column
	// ============================================================================
	case 1054:
		appErr := AutoSource(
			NewRepoError(ErrRepoDatabase,
				"unknown column"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("mysql_code", mysqlErr.Number).
			WithField("sql_state", mysqlErr.SQLState[0]).
			WithDetails(mysqlErr.Message)

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr

	// ============================================================================
	// 1205 — Lock Wait Timeout Exceeded
	// ============================================================================
	case 1205:
		appErr := AutoSource(
			NewRepoError(ErrRepoTimeout,
				"lock wait timeout"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("mysql_code", mysqlErr.Number).
			WithField("sql_state", mysqlErr.SQLState[0]).
			WithDetails("Lock wait timeout exceeded, transaction rolled back")

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr

	// ============================================================================
	// 1213 — Deadlock Found
	// ============================================================================
	case 1213:
		appErr := AutoSource(
			NewRepoError(ErrRepoDatabase,
				"deadlock detected"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("mysql_code", mysqlErr.Number).
			WithField("sql_state", mysqlErr.SQLState[0]).
			WithDetails("Deadlock found when trying to get lock, please retry transaction")

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr

	// ============================================================================
	// 1040 — Too Many Connections
	// ============================================================================
	case 1040:
		appErr := AutoSource(
			NewRepoError(ErrRepoConnection,
				"too many connections"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("mysql_code", mysqlErr.Number).
			WithField("sql_state", mysqlErr.SQLState[0]).
			WithDetails("Too many connections to MySQL server")

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr

	// ============================================================================
	// 1045 — Access Denied
	// ============================================================================
	case 1045:
		appErr := AutoSource(
			NewRepoError(ErrRepoDatabase,
				"access denied"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("mysql_code", mysqlErr.Number).
			WithField("sql_state", mysqlErr.SQLState[0]).
			WithDetails("Access denied for user")

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr

	// ============================================================================
	// 1044 — Access Denied for Database
	// ============================================================================
	case 1044:
		appErr := AutoSource(
			NewRepoError(ErrRepoDatabase,
				"access denied for database"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("mysql_code", mysqlErr.Number).
			WithField("sql_state", mysqlErr.SQLState[0]).
			WithDetails(mysqlErr.Message)

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr

	// ============================================================================
	// 1364 — Field Doesn't Have a Default Value
	// ============================================================================
	case 1364:
		appErr := AutoSource(
			NewRepoError(ErrRepoConstraint,
				"field has no default value"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("mysql_code", mysqlErr.Number).
			WithField("sql_state", mysqlErr.SQLState[0]).
			WithDetails(mysqlErr.Message)

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr

	// ============================================================================
	// 1406 — Data Too Long for Column
	// ============================================================================
	case 1406:
		appErr := AutoSource(
			NewRepoError(ErrRepoDatabase,
				"data too long"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("mysql_code", mysqlErr.Number).
			WithField("sql_state", mysqlErr.SQLState[0]).
			WithDetails("Data too long for column")

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr

	// ============================================================================
	// Default: Other MySQL errors
	// ============================================================================
	default:
		appErr := AutoSource(
			WrapRepoError(fmt.Errorf("mysql error: %s", mysqlErr.Message), ErrRepoDatabase,
				"mysql error"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("mysql_code", mysqlErr.Number).
			WithField("sql_state", mysqlErr.SQLState[0]).
			WithDetails(mysqlErr.Message)

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr
	}
}

// ============================================================================
// Error Detection Functions
// ============================================================================

func isMySQLConnectionError(msg string) bool {
	return strings.Contains(msg, "connection") ||
		strings.Contains(msg, "dial tcp") ||
		strings.Contains(msg, "connect") ||
		strings.Contains(msg, "EOF") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "can't connect to MySQL server")
}

func handleMySQLConnectionError(table string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoConnection,
			"mysql connection error"))

	if table != "" {
		_ = appErr.WithField("table", table)
	}
	_ = appErr.WithDetails("Failed to connect to MySQL server")

	for key, value := range extraFields {
		_ = appErr.WithField(key, value)
	}
	return appErr
}

// ============================================================================
// MySQL Error Codes Reference
// ============================================================================
// https://dev.mysql.com/doc/mysql-errors/8.0/en/server-error-reference.html
//
// Connection Errors:
//   1040 - ER_CON_COUNT_ERROR - Too many connections
//   1129 - ER_HOST_IS_BLOCKED - Host is blocked
//   2002 - CR_CONNECTION_ERROR - Can't connect to local MySQL server
//   2003 - CR_CONN_HOST_ERROR - Can't connect to MySQL server
//   2013 - CR_SERVER_LOST - Lost connection to MySQL server during query
//
// Authentication Errors:
//   1044 - ER_DBACCESS_DENIED_ERROR - Access denied for database
//   1045 - ER_ACCESS_DENIED_ERROR - Access denied for user
//
// Constraint Violations:
//   1048 - ER_BAD_NULL_ERROR - Column cannot be null
//   1062 - ER_DUP_ENTRY - Duplicate entry for key
//   1169 - ER_DUP_UNIQUE - Can't write; duplicate key in table
//   1216 - ER_NO_REFERENCED_ROW - Cannot add or update child row
//   1217 - ER_ROW_IS_REFERENCED - Cannot delete or update parent row
//   1451 - ER_ROW_IS_REFERENCED_2 - Cannot delete or update parent row (FK)
//   1452 - ER_NO_REFERENCED_ROW_2 - Cannot add or update child row (FK)
//   1557 - ER_FOREIGN_DUPLICATE_KEY - Upholding foreign key constraints
//
// Schema Errors:
//   1050 - ER_TABLE_EXISTS_ERROR - Table already exists
//   1051 - ER_BAD_TABLE_ERROR - Unknown table
//   1054 - ER_BAD_FIELD_ERROR - Unknown column
//   1060 - ER_DUP_FIELDNAME - Duplicate column name
//   1091 - ER_CANT_DROP_FIELD_OR_KEY - Can't DROP field/key; doesn't exist
//   1146 - ER_NO_SUCH_TABLE - Table doesn't exist
//   1364 - ER_NO_DEFAULT_FOR_FIELD - Field doesn't have a default value
//
// Data Errors:
//   1242 - ER_SUBQUERY_NO_1_ROW - Subquery returns more than 1 row
//   1264 - ER_WARN_DATA_OUT_OF_RANGE - Out of range value for column
//   1265 - ER_WARN_DATA_TRUNCATED - Data truncated for column
//   1292 - ER_TRUNCATED_WRONG_VALUE - Truncated incorrect value
//   1366 - ER_TRUNCATED_WRONG_VALUE_FOR_FIELD - Incorrect value for column
//   1406 - ER_DATA_TOO_LONG - Data too long for column
//
// Transaction Errors:
//   1205 - ER_LOCK_WAIT_TIMEOUT - Lock wait timeout exceeded
//   1213 - ER_LOCK_DEADLOCK - Deadlock found when trying to get lock
//   1614 - ER_XA_RBROLLBACK - Transaction branch was rolled back
//
// Syntax Errors:
//   1064 - ER_PARSE_ERROR - You have an error in your SQL syntax
//   1065 - ER_EMPTY_QUERY - Query was empty
//
// Operational Errors:
//   1030 - ER_FILSORT_ABORT - Got error from storage engine
//   1114 - ER_RECORD_FILE_FULL - The table is full
//   1203 - ER_TOO_MANY_USER_CONNECTIONS - User has too many connections
//   1226 - ER_USER_LIMIT_REACHED - User has exceeded resource limits
//
// Server Errors:
//   1006 - ER_CANT_CREATE_DB - Can't create database
//   1007 - ER_DB_CREATE_EXISTS - Can't create database; database exists
//   1008 - ER_DB_DROP_EXISTS - Can't drop database; database doesn't exist
//   1010 - ER_DB_DROP_DELETE - Error dropping database
