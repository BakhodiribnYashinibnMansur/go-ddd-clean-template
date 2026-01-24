package errors

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestHandlePgError_NoError(t *testing.T) {
	result := HandlePgError(nil, "users", nil)

	if result != nil {
		t.Errorf("HandlePgError(nil) should return nil, got %v", result)
	}
}

func TestHandlePgError_NoRows(t *testing.T) {
	result := HandlePgError(pgx.ErrNoRows, "users", map[string]any{
		"user_id": 123,
	})

	if result == nil {
		t.Fatal("HandlePgError(pgx.ErrNoRows) should return AppError")
	}

	if result.Type != ErrRepoNotFound {
		t.Errorf("Expected type %s, got %s", ErrRepoNotFound, result.Type)
	}

	if result.Fields["table"] != "users" {
		t.Errorf("Expected table=users, got %v", result.Fields["table"])
	}

	if result.Fields["user_id"] != 123 {
		t.Errorf("Expected user_id=123, got %v", result.Fields["user_id"])
	}
}

func TestHandlePgError_UniqueViolation(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:           "23505",
		Message:        "duplicate key value violates unique constraint",
		ConstraintName: "users_username_key",
		Severity:       "ERROR",
	}

	result := HandlePgError(pgErr, "users", map[string]any{
		"username": "john",
	})

	if result == nil {
		t.Fatal("HandlePgError should return AppError for unique violation")
	}

	if result.Type != ErrRepoAlreadyExists {
		t.Errorf("Expected type %s, got %s", ErrRepoAlreadyExists, result.Type)
	}

	if result.Fields["constraint"] != "users_username_key" {
		t.Errorf("Expected constraint=users_username_key, got %v", result.Fields["constraint"])
	}

	if result.Fields["pg_code"] != "23505" {
		t.Errorf("Expected pg_code=23505, got %v", result.Fields["pg_code"])
	}
}

func TestHandlePgError_ForeignKeyViolation(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:           "23503",
		Message:        "insert or update violates foreign key constraint",
		ConstraintName: "posts_user_id_fkey",
		Severity:       "ERROR",
	}

	result := HandlePgError(pgErr, "posts", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for FK violation")
	}

	if result.Type != ErrRepoConstraint {
		t.Errorf("Expected type %s, got %s", ErrRepoConstraint, result.Type)
	}
}

func TestHandlePgError_NotNullViolation(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:           "23502",
		Message:        "null value in column violates not-null constraint",
		ConstraintName: "users_email_not_null",
		ColumnName:     "email",
		Severity:       "ERROR",
	}

	result := HandlePgError(pgErr, "users", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for not-null violation")
	}

	if result.Type != ErrRepoConstraint {
		t.Errorf("Expected type %s, got %s", ErrRepoConstraint, result.Type)
	}

	if result.Fields["column"] != "email" {
		t.Errorf("Expected column=email, got %v", result.Fields["column"])
	}
}

func TestHandlePgError_CheckViolation(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:           "23514",
		Message:        "new row violates check constraint",
		ConstraintName: "users_age_check",
		Severity:       "ERROR",
	}

	result := HandlePgError(pgErr, "users", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for check violation")
	}

	if result.Type != ErrRepoConstraint {
		t.Errorf("Expected type %s, got %s", ErrRepoConstraint, result.Type)
	}
}

func TestHandlePgError_ConnectionError(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:     "08006",
		Message:  "connection failure",
		Severity: "FATAL",
	}

	result := HandlePgError(pgErr, "users", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for connection error")
	}

	if result.Type != ErrRepoConnection {
		t.Errorf("Expected type %s, got %s", ErrRepoConnection, result.Type)
	}
}

func TestHandlePgError_Deadlock(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:     "40P01",
		Message:  "deadlock detected",
		Severity: "ERROR",
	}

	result := HandlePgError(pgErr, "orders", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for deadlock")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}
}

func TestHandlePgError_LockTimeout(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:     "55P03",
		Message:  "lock not available",
		Severity: "ERROR",
	}

	result := HandlePgError(pgErr, "users", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for lock timeout")
	}

	if result.Type != ErrRepoTimeout {
		t.Errorf("Expected type %s, got %s", ErrRepoTimeout, result.Type)
	}
}

func TestHandlePgError_QueryCanceled(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:     "57014",
		Message:  "query canceled",
		Severity: "ERROR",
	}

	result := HandlePgError(pgErr, "users", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for query canceled")
	}

	if result.Type != ErrRepoTimeout {
		t.Errorf("Expected type %s, got %s", ErrRepoTimeout, result.Type)
	}
}

func TestHandlePgError_GenericError(t *testing.T) {
	genericErr := errors.New("some database error")

	result := HandlePgError(genericErr, "users", map[string]any{
		"operation": "insert",
	})

	if result == nil {
		t.Fatal("HandlePgError should return AppError for generic error")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}

	if result.Fields["operation"] != "insert" {
		t.Errorf("Expected operation=insert, got %v", result.Fields["operation"])
	}
}

func TestHandlePgError_ExtraFields(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:           "23505",
		Message:        "duplicate key",
		ConstraintName: "users_email_key",
		Severity:       "ERROR",
	}

	extraFields := map[string]any{
		"user_id":   123,
		"email":     "test@example.com",
		"operation": "create_user",
	}

	result := HandlePgError(pgErr, "users", extraFields)

	if result == nil {
		t.Fatal("HandlePgError should return AppError")
	}

	// Check all extra fields are present
	for key, expectedValue := range extraFields {
		if result.Fields[key] != expectedValue {
			t.Errorf("Expected field %s=%v, got %v", key, expectedValue, result.Fields[key])
		}
	}
}

func TestHandlePgError_DataException(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:     "22003",
		Message:  "numeric value out of range",
		Severity: "ERROR",
	}

	result := HandlePgError(pgErr, "products", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for data exception")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}
}

func TestHandlePgError_TransactionError(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:     "25P02",
		Message:  "in failed sql transaction",
		Severity: "ERROR",
	}

	result := HandlePgError(pgErr, "users", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for transaction error")
	}

	if result.Type != ErrRepoTransaction {
		t.Errorf("Expected type %s, got %s", ErrRepoTransaction, result.Type)
	}
}

func TestHandlePgError_AuthError(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:     "28P01",
		Message:  "invalid password",
		Severity: "FATAL",
	}

	result := HandlePgError(pgErr, "", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for auth error")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}
}

func TestHandlePgError_SyntaxError(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:     "42601",
		Message:  "syntax error at or near",
		Severity: "ERROR",
	}

	result := HandlePgError(pgErr, "users", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for syntax error")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}
}

func TestHandlePgError_InsufficientPrivilege(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:     "42501",
		Message:  "permission denied",
		Severity: "ERROR",
	}

	result := HandlePgError(pgErr, "users", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for permission error")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}
}

func TestHandlePgError_TableNotFound(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:     "42P01",
		Message:  "relation does not exist",
		Severity: "ERROR",
	}

	result := HandlePgError(pgErr, "nonexistent", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for table not found")
	}

	if result.Type != ErrRepoDatabase {
		t.Errorf("Expected type %s, got %s", ErrRepoDatabase, result.Type)
	}
}

func TestHandlePgError_EmptyTable(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:           "23505",
		Message:        "duplicate key",
		ConstraintName: "pk",
		Severity:       "ERROR",
	}

	result := HandlePgError(pgErr, "", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError")
	}

	// Should work without table name
	if result.Type != ErrRepoAlreadyExists {
		t.Errorf("Expected type %s, got %s", ErrRepoAlreadyExists, result.Type)
	}
}

func TestHandlePgError_NilExtraFields(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:     "23505",
		Message:  "duplicate",
		Severity: "ERROR",
	}

	result := HandlePgError(pgErr, "users", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError")
	}

	// Should work with nil extra fields
	if result.Type != ErrRepoAlreadyExists {
		t.Errorf("Expected type %s, got %s", ErrRepoAlreadyExists, result.Type)
	}
}

// Benchmark tests
func BenchmarkHandlePgError_NoRows(b *testing.B) {
	for range b.N {
		HandlePgError(pgx.ErrNoRows, "users", nil)
	}
}

func BenchmarkHandlePgError_UniqueViolation(b *testing.B) {
	pgErr := &pgconn.PgError{
		Code:           "23505",
		Message:        "duplicate key",
		ConstraintName: "users_email_key",
		Severity:       "ERROR",
	}

	b.ResetTimer()
	for range b.N {
		HandlePgError(pgErr, "users", nil)
	}
}

func BenchmarkHandlePgError_GenericError(b *testing.B) {
	err := errors.New("database error")

	b.ResetTimer()
	for range b.N {
		HandlePgError(err, "users", nil)
	}
}
