package errorx

import (
	"github.com/jackc/pgx/v5/pgconn"
)

// handleAuthError handles Class 28 errors (Authorization)
func handleAuthError(pgErr *pgconn.PgError, table string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoDatabase,
			"authentication failed"))

	if table != "" {
		_ = appErr.WithField("table", table)
	}
	_ = appErr.WithField("pg_code", pgErr.Code).
		WithField("pg_severity", pgErr.Severity).
		WithDetails(pgErr.Message)

	for key, value := range extraFields {
		_ = appErr.WithField(key, value)
	}
	return appErr
}

// handleSyntaxOrAccessError handles Class 42 errors
func handleSyntaxOrAccessError(pgErr *pgconn.PgError, table string, extraFields map[string]any) *AppError {
	// 42501 is insufficient_privilege
	if pgErr.Code == "42501" {
		appErr := AutoSource(
			NewRepoError(ErrRepoDatabase,
				"insufficient privilege"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("pg_code", pgErr.Code).
			WithDetails("User does not have required database privileges")

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr
	}

	// 42P01 is undefined_table
	if pgErr.Code == "42P01" {
		appErr := AutoSource(
			NewRepoError(ErrRepoDatabase,
				"table does not exist"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("pg_code", pgErr.Code).
			WithDetails(pgErr.Message)

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr
	}

	appErr := AutoSource(
		NewRepoError(ErrRepoDatabase,
			"syntax error or access violation"))

	if table != "" {
		_ = appErr.WithField("table", table)
	}
	_ = appErr.WithField("pg_code", pgErr.Code).
		WithField("pg_severity", pgErr.Severity).
		WithDetails(pgErr.Message)

	for key, value := range extraFields {
		_ = appErr.WithField(key, value)
	}
	return appErr
}

// handleConfigError handles Class F0 errors
func handleConfigError(pgErr *pgconn.PgError, table string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoDatabase,
			"configuration file error"))

	if table != "" {
		_ = appErr.WithField("table", table)
	}
	_ = appErr.WithField("pg_code", pgErr.Code).
		WithField("pg_severity", pgErr.Severity).
		WithDetails(pgErr.Message)

	for key, value := range extraFields {
		_ = appErr.WithField(key, value)
	}
	return appErr
}

// handleFDWError handles Class HV errors (Foreign Data Wrapper)
func handleFDWError(pgErr *pgconn.PgError, table string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoDatabase,
			"foreign data wrapper error"))

	if table != "" {
		_ = appErr.WithField("table", table)
	}
	_ = appErr.WithField("pg_code", pgErr.Code).
		WithField("pg_severity", pgErr.Severity).
		WithDetails(pgErr.Message)

	for key, value := range extraFields {
		_ = appErr.WithField(key, value)
	}
	return appErr
}

// handlePLpgSQLError handles Class P0 errors
func handlePLpgSQLError(pgErr *pgconn.PgError, table string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoDatabase,
			"PL/pgSQL error"))

	if table != "" {
		_ = appErr.WithField("table", table)
	}
	_ = appErr.WithField("pg_code", pgErr.Code).
		WithField("pg_severity", pgErr.Severity).
		WithDetails(pgErr.Message)

	for key, value := range extraFields {
		_ = appErr.WithField(key, value)
	}
	return appErr
}

// handleInternalError handles Class XX errors
func handleInternalError(pgErr *pgconn.PgError, table string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoDatabase,
			"internal database error"))

	if table != "" {
		_ = appErr.WithField("table", table)
	}
	_ = appErr.WithField("pg_code", pgErr.Code).
		WithField("pg_severity", pgErr.Severity).
		WithDetails(pgErr.Message)

	for key, value := range extraFields {
		_ = appErr.WithField(key, value)
	}
	return appErr
}
