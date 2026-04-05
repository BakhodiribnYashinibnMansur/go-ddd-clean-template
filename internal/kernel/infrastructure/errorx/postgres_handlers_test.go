package errorx

import (
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestHandleAuthError(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:     "28P01",
		Severity: "FATAL",
		Message:  "password authentication failed",
	}

	appErr := handleAuthError(pgErr, "users", map[string]any{"operation": "select"})

	if appErr == nil {
		t.Fatal("expected non-nil error")
	}
	if appErr.Type != ErrRepoDatabase {
		t.Errorf("expected type %s, got %s", ErrRepoDatabase, appErr.Type)
	}
	if appErr.Fields["table"] != "users" {
		t.Errorf("expected table 'users', got %v", appErr.Fields["table"])
	}
	if appErr.Fields["pg_code"] != "28P01" {
		t.Errorf("expected pg_code '28P01', got %v", appErr.Fields["pg_code"])
	}
	if appErr.Fields["operation"] != "select" {
		t.Errorf("expected operation 'select', got %v", appErr.Fields["operation"])
	}
}

func TestHandleAuthError_EmptyTable(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:     "28P01",
		Severity: "FATAL",
		Message:  "auth failed",
	}

	appErr := handleAuthError(pgErr, "", nil)

	if appErr == nil {
		t.Fatal("expected non-nil error")
	}
	if _, ok := appErr.Fields["table"]; ok {
		t.Error("expected no table field for empty table")
	}
}

func TestHandleSyntaxOrAccessError(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		wantMsg    string
		table      string
		extraField map[string]any
	}{
		{
			name:    "insufficient privilege",
			code:    "42501",
			wantMsg: "insufficient privilege",
			table:   "users",
		},
		{
			name:    "undefined table",
			code:    "42P01",
			wantMsg: "table does not exist",
			table:   "unknown_table",
		},
		{
			name:    "generic syntax error",
			code:    "42601",
			wantMsg: "syntax error or access violation",
			table:   "orders",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pgErr := &pgconn.PgError{
				Code:     tt.code,
				Severity: "ERROR",
				Message:  "test pg message",
			}

			appErr := handleSyntaxOrAccessError(pgErr, tt.table, tt.extraField)

			if appErr == nil {
				t.Fatal("expected non-nil error")
			}
			if appErr.Type != ErrRepoDatabase {
				t.Errorf("expected type %s, got %s", ErrRepoDatabase, appErr.Type)
			}
			if appErr.Message != tt.wantMsg {
				t.Errorf("expected message %q, got %q", tt.wantMsg, appErr.Message)
			}
			if appErr.Fields["table"] != tt.table {
				t.Errorf("expected table %q, got %v", tt.table, appErr.Fields["table"])
			}
		})
	}
}

func TestHandleSyntaxOrAccessError_WithExtraFields(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:     "42601",
		Severity: "ERROR",
		Message:  "syntax error at position 15",
	}

	extra := map[string]any{
		"query":     "SELECT * FROM ...",
		"operation": "list",
	}

	appErr := handleSyntaxOrAccessError(pgErr, "users", extra)

	if appErr.Fields["query"] != "SELECT * FROM ..." {
		t.Errorf("expected query field, got %v", appErr.Fields["query"])
	}
	if appErr.Fields["operation"] != "list" {
		t.Errorf("expected operation field, got %v", appErr.Fields["operation"])
	}
}

func TestHandleConfigError(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:     "F0001",
		Severity: "FATAL",
		Message:  "lock file exists",
	}

	appErr := handleConfigError(pgErr, "config_table", nil)

	if appErr == nil {
		t.Fatal("expected non-nil error")
	}
	if appErr.Type != ErrRepoDatabase {
		t.Errorf("expected type %s, got %s", ErrRepoDatabase, appErr.Type)
	}
	if appErr.Fields["pg_code"] != "F0001" {
		t.Errorf("expected pg_code 'F0001', got %v", appErr.Fields["pg_code"])
	}
}

func TestHandleFDWError(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:     "HV000",
		Severity: "ERROR",
		Message:  "fdw error",
	}

	appErr := handleFDWError(pgErr, "foreign_table", map[string]any{"key": "val"})

	if appErr == nil {
		t.Fatal("expected non-nil error")
	}
	if appErr.Type != ErrRepoDatabase {
		t.Errorf("expected type %s, got %s", ErrRepoDatabase, appErr.Type)
	}
	if appErr.Fields["key"] != "val" {
		t.Errorf("expected extra field 'key'='val', got %v", appErr.Fields["key"])
	}
}

func TestHandlePLpgSQLError(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:     "P0001",
		Severity: "ERROR",
		Message:  "raise exception",
	}

	appErr := handlePLpgSQLError(pgErr, "proc_table", nil)

	if appErr == nil {
		t.Fatal("expected non-nil error")
	}
	if appErr.Type != ErrRepoDatabase {
		t.Errorf("expected type %s, got %s", ErrRepoDatabase, appErr.Type)
	}
	if appErr.Details != "raise exception" {
		t.Errorf("expected details 'raise exception', got %q", appErr.Details)
	}
}

func TestHandleInternalError(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:     "XX001",
		Severity: "FATAL",
		Message:  "data corrupted",
	}

	appErr := handleInternalError(pgErr, "data_table", nil)

	if appErr == nil {
		t.Fatal("expected non-nil error")
	}
	if appErr.Type != ErrRepoDatabase {
		t.Errorf("expected type %s, got %s", ErrRepoDatabase, appErr.Type)
	}
	if appErr.Fields["pg_code"] != "XX001" {
		t.Errorf("expected pg_code 'XX001', got %v", appErr.Fields["pg_code"])
	}
	if appErr.Fields["pg_severity"] != "FATAL" {
		t.Errorf("expected pg_severity 'FATAL', got %v", appErr.Fields["pg_severity"])
	}
}
