package errors

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestHandlePgError_NoError(t *testing.T) {
	ctx := t.Context()
	result := HandlePgError(ctx, nil, "users", nil)

	if result != nil {
		t.Errorf("HandlePgError(nil) should return nil, got %v", result)
	}
}

func TestHandlePgError_NoRows(t *testing.T) {
	ctx := t.Context()
	result := HandlePgError(ctx, pgx.ErrNoRows, "users", map[string]any{
		"user_id": 123,
	})

	if result == nil {
		t.Fatal("HandlePgError(pgx.ErrNoRows) should return AppError")
	}

	if result.Code != ErrRepoNotFound {
		t.Errorf("Expected code %s, got %s", ErrRepoNotFound, result.Code)
	}

	if result.Fields["table"] != "users" {
		t.Errorf("Expected table=users, got %v", result.Fields["table"])
	}

	if result.Fields["user_id"] != 123 {
		t.Errorf("Expected user_id=123, got %v", result.Fields["user_id"])
	}
}

func TestHandlePgError_UniqueViolation(t *testing.T) {
	ctx := t.Context()
	pgErr := &pgconn.PgError{
		Code:           "23505",
		Message:        "duplicate key value violates unique constraint",
		ConstraintName: "users_username_key",
		Severity:       "ERROR",
	}

	result := HandlePgError(ctx, pgErr, "users", map[string]any{
		"username": "john",
	})

	if result == nil {
		t.Fatal("HandlePgError should return AppError for unique violation")
	}

	if result.Code != ErrRepoAlreadyExists {
		t.Errorf("Expected code %s, got %s", ErrRepoAlreadyExists, result.Code)
	}

	if result.Fields["constraint"] != "users_username_key" {
		t.Errorf("Expected constraint=users_username_key, got %v", result.Fields["constraint"])
	}

	if result.Fields["pg_code"] != "23505" {
		t.Errorf("Expected pg_code=23505, got %v", result.Fields["pg_code"])
	}
}

func TestHandlePgError_ForeignKeyViolation(t *testing.T) {
	ctx := t.Context()
	pgErr := &pgconn.PgError{
		Code:           "23503",
		Message:        "insert or update violates foreign key constraint",
		ConstraintName: "posts_user_id_fkey",
		Severity:       "ERROR",
	}

	result := HandlePgError(ctx, pgErr, "posts", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for FK violation")
	}

	if result.Code != ErrRepoConstraint {
		t.Errorf("Expected code %s, got %s", ErrRepoConstraint, result.Code)
	}
}

func TestHandlePgError_NotNullViolation(t *testing.T) {
	ctx := t.Context()
	pgErr := &pgconn.PgError{
		Code:           "23502",
		Message:        "null value in column violates not-null constraint",
		ConstraintName: "users_email_not_null",
		ColumnName:     "email",
		Severity:       "ERROR",
	}

	result := HandlePgError(ctx, pgErr, "users", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for not-null violation")
	}

	if result.Code != ErrRepoConstraint {
		t.Errorf("Expected code %s, got %s", ErrRepoConstraint, result.Code)
	}

	if result.Fields["column"] != "email" {
		t.Errorf("Expected column=email, got %v", result.Fields["column"])
	}
}

func TestHandlePgError_CheckViolation(t *testing.T) {
	ctx := t.Context()
	pgErr := &pgconn.PgError{
		Code:           "23514",
		Message:        "new row violates check constraint",
		ConstraintName: "users_age_check",
		Severity:       "ERROR",
	}

	result := HandlePgError(ctx, pgErr, "users", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for check violation")
	}

	if result.Code != ErrRepoConstraint {
		t.Errorf("Expected code %s, got %s", ErrRepoConstraint, result.Code)
	}
}

func TestHandlePgError_ConnectionError(t *testing.T) {
	ctx := t.Context()
	pgErr := &pgconn.PgError{
		Code:     "08006",
		Message:  "connection failure",
		Severity: "FATAL",
	}

	result := HandlePgError(ctx, pgErr, "users", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for connection error")
	}

	if result.Code != ErrRepoConnection {
		t.Errorf("Expected code %s, got %s", ErrRepoConnection, result.Code)
	}
}

func TestHandlePgError_Deadlock(t *testing.T) {
	ctx := t.Context()
	pgErr := &pgconn.PgError{
		Code:     "40P01",
		Message:  "deadlock detected",
		Severity: "ERROR",
	}

	result := HandlePgError(ctx, pgErr, "orders", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for deadlock")
	}

	if result.Code != ErrRepoDatabase {
		t.Errorf("Expected code %s, got %s", ErrRepoDatabase, result.Code)
	}
}

func TestHandlePgError_LockTimeout(t *testing.T) {
	ctx := t.Context()
	pgErr := &pgconn.PgError{
		Code:     "55P03",
		Message:  "lock not available",
		Severity: "ERROR",
	}

	result := HandlePgError(ctx, pgErr, "users", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for lock timeout")
	}

	if result.Code != ErrRepoTimeout {
		t.Errorf("Expected code %s, got %s", ErrRepoTimeout, result.Code)
	}
}

func TestHandlePgError_QueryCanceled(t *testing.T) {
	ctx := t.Context()
	pgErr := &pgconn.PgError{
		Code:     "57014",
		Message:  "query canceled",
		Severity: "ERROR",
	}

	result := HandlePgError(ctx, pgErr, "users", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for query canceled")
	}

	if result.Code != ErrRepoTimeout {
		t.Errorf("Expected code %s, got %s", ErrRepoTimeout, result.Code)
	}
}

func TestHandlePgError_GenericError(t *testing.T) {
	ctx := t.Context()
	genericErr := errors.New("some database error")

	result := HandlePgError(ctx, genericErr, "users", map[string]any{
		"operation": "insert",
	})

	if result == nil {
		t.Fatal("HandlePgError should return AppError for generic error")
	}

	if result.Code != ErrRepoDatabase {
		t.Errorf("Expected code %s, got %s", ErrRepoDatabase, result.Code)
	}

	if result.Fields["operation"] != "insert" {
		t.Errorf("Expected operation=insert, got %v", result.Fields["operation"])
	}
}

func TestHandlePgError_ExtraFields(t *testing.T) {
	ctx := t.Context()
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

	result := HandlePgError(ctx, pgErr, "users", extraFields)

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
	ctx := t.Context()
	pgErr := &pgconn.PgError{
		Code:     "22003",
		Message:  "numeric value out of range",
		Severity: "ERROR",
	}

	result := HandlePgError(ctx, pgErr, "products", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for data exception")
	}

	if result.Code != ErrRepoDatabase {
		t.Errorf("Expected code %s, got %s", ErrRepoDatabase, result.Code)
	}
}

func TestHandlePgError_TransactionError(t *testing.T) {
	ctx := t.Context()
	pgErr := &pgconn.PgError{
		Code:     "25P02",
		Message:  "in failed sql transaction",
		Severity: "ERROR",
	}

	result := HandlePgError(ctx, pgErr, "users", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for transaction error")
	}

	if result.Code != ErrRepoTransaction {
		t.Errorf("Expected code %s, got %s", ErrRepoTransaction, result.Code)
	}
}

func TestHandlePgError_AuthError(t *testing.T) {
	ctx := t.Context()
	pgErr := &pgconn.PgError{
		Code:     "28P01",
		Message:  "invalid password",
		Severity: "FATAL",
	}

	result := HandlePgError(ctx, pgErr, "", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for auth error")
	}

	if result.Code != ErrRepoDatabase {
		t.Errorf("Expected code %s, got %s", ErrRepoDatabase, result.Code)
	}
}

func TestHandlePgError_SyntaxError(t *testing.T) {
	ctx := t.Context()
	pgErr := &pgconn.PgError{
		Code:     "42601",
		Message:  "syntax error at or near",
		Severity: "ERROR",
	}

	result := HandlePgError(ctx, pgErr, "users", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for syntax error")
	}

	if result.Code != ErrRepoDatabase {
		t.Errorf("Expected code %s, got %s", ErrRepoDatabase, result.Code)
	}
}

func TestHandlePgError_InsufficientPrivilege(t *testing.T) {
	ctx := t.Context()
	pgErr := &pgconn.PgError{
		Code:     "42501",
		Message:  "permission denied",
		Severity: "ERROR",
	}

	result := HandlePgError(ctx, pgErr, "users", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for permission error")
	}

	if result.Code != ErrRepoDatabase {
		t.Errorf("Expected code %s, got %s", ErrRepoDatabase, result.Code)
	}
}

func TestHandlePgError_TableNotFound(t *testing.T) {
	ctx := t.Context()
	pgErr := &pgconn.PgError{
		Code:     "42P01",
		Message:  "relation does not exist",
		Severity: "ERROR",
	}

	result := HandlePgError(ctx, pgErr, "nonexistent", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError for table not found")
	}

	if result.Code != ErrRepoDatabase {
		t.Errorf("Expected code %s, got %s", ErrRepoDatabase, result.Code)
	}
}

func TestHandlePgError_EmptyTable(t *testing.T) {
	ctx := t.Context()
	pgErr := &pgconn.PgError{
		Code:           "23505",
		Message:        "duplicate key",
		ConstraintName: "pk",
		Severity:       "ERROR",
	}

	result := HandlePgError(ctx, pgErr, "", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError")
	}

	// Should work without table name
	if result.Code != ErrRepoAlreadyExists {
		t.Errorf("Expected code %s, got %s", ErrRepoAlreadyExists, result.Code)
	}
}

func TestHandlePgError_NilExtraFields(t *testing.T) {
	ctx := t.Context()
	pgErr := &pgconn.PgError{
		Code:     "23505",
		Message:  "duplicate",
		Severity: "ERROR",
	}

	result := HandlePgError(ctx, pgErr, "users", nil)

	if result == nil {
		t.Fatal("HandlePgError should return AppError")
	}

	// Should work with nil extra fields
	if result.Code != ErrRepoAlreadyExists {
		t.Errorf("Expected code %s, got %s", ErrRepoAlreadyExists, result.Code)
	}
}

// Benchmark tests
func BenchmarkHandlePgError_NoRows(b *testing.B) {
	ctx := b.Context()
	for range b.N {
		HandlePgError(ctx, pgx.ErrNoRows, "users", nil)
	}
}

func BenchmarkHandlePgError_UniqueViolation(b *testing.B) {
	ctx := b.Context()
	pgErr := &pgconn.PgError{
		Code:           "23505",
		Message:        "duplicate key",
		ConstraintName: "users_email_key",
		Severity:       "ERROR",
	}

	b.ResetTimer()
	for range b.N {
		HandlePgError(ctx, pgErr, "users", nil)
	}
}

func BenchmarkHandlePgError_GenericError(b *testing.B) {
	ctx := b.Context()
	err := errors.New("database error")

	b.ResetTimer()
	for range b.N {
		HandlePgError(ctx, err, "users", nil)
	}
}
