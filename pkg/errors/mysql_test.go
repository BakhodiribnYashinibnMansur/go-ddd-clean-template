package errors

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestHandleMySQLError_NoError(t *testing.T) {
	result := HandleMySQLError(nil, "users", nil)

	if result != nil {
		t.Errorf("HandleMySQLError(nil) should return nil, got %v", result)
	}
}

func TestHandleMySQLError_NoRows(t *testing.T) {
	result := HandleMySQLError(sql.ErrNoRows, "users", map[string]any{
		"user_id": 123,
	})

	if result == nil {
		t.Fatal("HandleMySQLError(sql.ErrNoRows) should return AppError")
	}

	if result.Type != ErrRepoNotFound {
		t.Errorf("Expected type %s, got %s", ErrRepoNotFound, result.Type)
	}

	if result.Fields["table"] != "users" {
		t.Errorf("Expected table=users, got %v", result.Fields["table"])
	}
}

func TestHandleMySQLError_DuplicateEntry(t *testing.T) {
	mysqlErr := &mysql.MySQLError{
		Number:   1062,
		SQLState: [5]byte{'2', '3', '0', '0', '0'},
		Message:  "Duplicate entry 'john' for key 'username'",
	}

	result := HandleMySQLError(mysqlErr, "users", map[string]any{
		"username": "john",
	})

	if result == nil {
		t.Fatal("HandleMySQLError should return AppError for duplicate entry")
	}

	if result.Type != ErrRepoAlreadyExists {
		t.Errorf("Expected type %s, got %s", ErrRepoAlreadyExists, result.Type)
	}

	if result.Fields["mysql_code"] != uint16(1062) {
		t.Errorf("Expected mysql_code=1062, got %v", result.Fields["mysql_code"])
	}
}

func TestHandleMySQLError_ForeignKeyConstraint(t *testing.T) {
	mysqlErr := &mysql.MySQLError{
		Number:   1452,
		SQLState: [5]byte{'2', '3', '0', '0', '0'},
		Message:  "Cannot add or update child row: a foreign key constraint fails",
	}

	result := HandleMySQLError(mysqlErr, "posts", nil)

	if result == nil {
		t.Fatal("HandleMySQLError should return AppError for FK constraint")
	}

	if result.Type != ErrRepoConstraint {
		t.Errorf("Expected type %s, got %s", ErrRepoConstraint, result.Type)
	}
}

func TestHandleMySQLError_CannotDeleteParentRow(t *testing.T) {
	mysqlErr := &mysql.MySQLError{
		Number:   1451,
		SQLState: [5]byte{'2', '3', '0', '0', '0'},
		Message:  "Cannot delete or update a parent row: a foreign key constraint fails",
	}

	result := HandleMySQLError(mysqlErr, "users", nil)

	if result == nil {
		t.Fatal("HandleMySQLError should return AppError")
	}

	if result.Type != ErrRepoConstraint {
		t.Errorf("Expected type %s, got %s", ErrRepoConstraint, result.Type)
	}
}

func TestHandleMySQLError_ColumnCannotBeNull(t *testing.T) {
	mysqlErr := &mysql.MySQLError{
		Number:   1048,
		SQLState: [5]byte{'2', '3', '0', '0', '0'},
		Message:  "Column 'email' cannot be null",
	}

	result := HandleMySQLError(mysqlErr, "users", nil)

	if result == nil {
		t.Fatal("HandleMySQLError should return AppError for not null")
	}

	if result.Type != ErrRepoConstraint {
		t.Errorf("Expected type %s, got %s", ErrRepoConstraint, result.Type)
	}
}

func TestHandleMySQLError_TableDoesntExist(t *testing.T) {
	mysqlErr := &mysql.MySQLError{
		Number:   1146,
		SQLState: [5]byte{'4', '2', 'S', '0', '2'},
		Message:  "Table 'db.nonexistent' doesn't exist",
	}

	result := HandleMySQLError(mysqlErr, "nonexistent", nil)

	if result == nil {
		t.Fatal("HandleMySQLError should return AppError")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}
}

func TestHandleMySQLError_UnknownColumn(t *testing.T) {
	mysqlErr := &mysql.MySQLError{
		Number:   1054,
		SQLState: [5]byte{'4', '2', 'S', '2', '2'},
		Message:  "Unknown column 'xyz' in 'field list'",
	}

	result := HandleMySQLError(mysqlErr, "users", nil)

	if result == nil {
		t.Fatal("HandleMySQLError should return AppError")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}
}

func TestHandleMySQLError_LockWaitTimeout(t *testing.T) {
	mysqlErr := &mysql.MySQLError{
		Number:   1205,
		SQLState: [5]byte{'H', 'Y', '0', '0', '0'},
		Message:  "Lock wait timeout exceeded; try restarting transaction",
	}

	result := HandleMySQLError(mysqlErr, "orders", nil)

	if result == nil {
		t.Fatal("HandleMySQLError should return AppError for lock timeout")
	}

	if result.Type != ErrRepoTimeout {
		t.Errorf("Expected type %s, got %s", ErrRepoTimeout, result.Type)
	}
}

func TestHandleMySQLError_Deadlock(t *testing.T) {
	mysqlErr := &mysql.MySQLError{
		Number:   1213,
		SQLState: [5]byte{'4', '0', '0', '0', '1'},
		Message:  "Deadlock found when trying to get lock; try restarting transaction",
	}

	result := HandleMySQLError(mysqlErr, "transfers", nil)

	if result == nil {
		t.Fatal("HandleMySQLError should return AppError for deadlock")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}
}

func TestHandleMySQLError_TooManyConnections(t *testing.T) {
	mysqlErr := &mysql.MySQLError{
		Number:   1040,
		SQLState: [5]byte{'0', '8', '0', '0', '4'},
		Message:  "Too many connections",
	}

	result := HandleMySQLError(mysqlErr, "", nil)

	if result == nil {
		t.Fatal("HandleMySQLError should return AppError for too many connections")
	}

	if result.Type != ErrRepoConnection {
		t.Errorf("Expected type %s, got %s", ErrRepoConnection, result.Type)
	}
}

func TestHandleMySQLError_AccessDenied(t *testing.T) {
	mysqlErr := &mysql.MySQLError{
		Number:   1045,
		SQLState: [5]byte{'2', '8', '0', '0', '0'},
		Message:  "Access denied for user 'user'@'localhost'",
	}

	result := HandleMySQLError(mysqlErr, "", nil)

	if result == nil {
		t.Fatal("HandleMySQLError should return AppError for access denied")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}
}

func TestHandleMySQLError_AccessDeniedForDatabase(t *testing.T) {
	mysqlErr := &mysql.MySQLError{
		Number:   1044,
		SQLState: [5]byte{'4', '2', '0', '0', '0'},
		Message:  "Access denied for user 'user'@'localhost' to database 'db'",
	}

	result := HandleMySQLError(mysqlErr, "", nil)

	if result == nil {
		t.Fatal("HandleMySQLError should return AppError")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}
}

func TestHandleMySQLError_NoDefaultValue(t *testing.T) {
	mysqlErr := &mysql.MySQLError{
		Number:   1364,
		SQLState: [5]byte{'H', 'Y', '0', '0', '0'},
		Message:  "Field 'created_at' doesn't have a default value",
	}

	result := HandleMySQLError(mysqlErr, "users", nil)

	if result == nil {
		t.Fatal("HandleMySQLError should return AppError")
	}

	if result.Type != ErrRepoConstraint {
		t.Errorf("Expected type %s, got %s", ErrRepoConstraint, result.Type)
	}
}

func TestHandleMySQLError_DataTooLong(t *testing.T) {
	mysqlErr := &mysql.MySQLError{
		Number:   1406,
		SQLState: [5]byte{'2', '2', '0', '0', '1'},
		Message:  "Data too long for column 'username' at row 1",
	}

	result := HandleMySQLError(mysqlErr, "users", nil)

	if result == nil {
		t.Fatal("HandleMySQLError should return AppError")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}
}

func TestHandleMySQLError_ConnectionError(t *testing.T) {
	err := errors.New("dial tcp: connection refused")

	result := HandleMySQLError(err, "users", nil)

	if result == nil {
		t.Fatal("HandleMySQLError should return AppError for connection error")
	}

	if result.Type != ErrRepoConnection {
		t.Errorf("Expected type %s, got %s", ErrRepoConnection, result.Type)
	}
}

func TestHandleMySQLError_GenericError(t *testing.T) {
	err := errors.New("some database error")

	result := HandleMySQLError(err, "users", map[string]any{
		"operation": "insert",
	})

	if result == nil {
		t.Fatal("HandleMySQLError should return AppError for generic error")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}

	if result.Fields["operation"] != "insert" {
		t.Errorf("Expected operation=insert, got %v", result.Fields["operation"])
	}
}

func TestHandleMySQLError_ExtraFields(t *testing.T) {
	mysqlErr := &mysql.MySQLError{
		Number:   1062,
		SQLState: [5]byte{'2', '3', '0', '0', '0'},
		Message:  "Duplicate entry",
	}

	extraFields := map[string]any{
		"user_id":   123,
		"email":     "test@example.com",
		"operation": "create_user",
	}

	result := HandleMySQLError(mysqlErr, "users", extraFields)

	if result == nil {
		t.Fatal("HandleMySQLError should return AppError")
	}

	// Check all extra fields are present
	for key, expectedValue := range extraFields {
		if result.Fields[key] != expectedValue {
			t.Errorf("Expected field %s=%v, got %v", key, expectedValue, result.Fields[key])
		}
	}
}

func TestHandleMySQLError_UnknownMySQLError(t *testing.T) {
	mysqlErr := &mysql.MySQLError{
		Number:   9999,
		SQLState: [5]byte{'H', 'Y', '0', '0', '0'},
		Message:  "Unknown error",
	}

	result := HandleMySQLError(mysqlErr, "users", nil)

	if result == nil {
		t.Fatal("HandleMySQLError should return AppError")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}
}

// Benchmark tests
func BenchmarkHandleMySQLError_NoRows(b *testing.B) {
	for range b.N {
		HandleMySQLError(sql.ErrNoRows, "users", nil)
	}
}

func BenchmarkHandleMySQLError_DuplicateEntry(b *testing.B) {
	mysqlErr := &mysql.MySQLError{
		Number:   1062,
		SQLState: [5]byte{'2', '3', '0', '0', '0'},
		Message:  "Duplicate entry",
	}

	b.ResetTimer()
	for range b.N {
		HandleMySQLError(mysqlErr, "users", nil)
	}
}

func BenchmarkHandleMySQLError_GenericError(b *testing.B) {
	err := errors.New("database error")

	b.ResetTimer()
	for range b.N {
		HandleMySQLError(err, "users", nil)
	}
}
