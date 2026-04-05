package errorx

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// HandlePgError handles PostgreSQL errors and converts them to AppError
// This centralizes all PostgreSQL error handling logic
func HandlePgError(err error, table string, extraFields map[string]any) *AppError {
	if err == nil {
		return nil
	}

	// Check for pgx.ErrNoRows (special case)
	if errors.Is(err, pgx.ErrNoRows) {
		appErr := AutoSource(
			NewRepoError(ErrRepoNotFound,
				"record not found"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithDetails("No matching record found in database")

		// Add extra fields
		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}

		return appErr
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		// Not a PostgreSQL error, return generic database error
		appErr := AutoSource(
			WrapRepoError(err, ErrRepoDatabase,
				"database operation failed"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}

		// Add extra fields
		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}

		return appErr
	}

	// Handle specific PostgreSQL error codes by class
	switch {
	// ============================================================================
	// Class 08 — Connection Exception
	// ============================================================================
	case strings.HasPrefix(pgErr.Code, "08"):
		return handleConnectionError(pgErr, table, extraFields)

	// ============================================================================
	// Class 22 — Data Exception
	// ============================================================================
	case strings.HasPrefix(pgErr.Code, "22"):
		return handleDataException(pgErr, table, extraFields)

	// ============================================================================
	// Class 23 — Integrity Constraint Violation
	// ============================================================================
	case pgErr.Code == "23505": // unique_violation
		appErr := AutoSource(
			NewRepoError(ErrRepoAlreadyExists,
				"record already exists"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("constraint", pgErr.ConstraintName).
			WithField("pg_code", pgErr.Code).
			WithDetails("A record with this unique constraint already exists")

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr

	case pgErr.Code == "23503": // foreign_key_violation
		appErr := AutoSource(
			NewRepoError(ErrRepoConstraint,
				"foreign key constraint violation"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("constraint", pgErr.ConstraintName).
			WithField("pg_code", pgErr.Code).
			WithDetails(pgErr.Message)

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr

	case pgErr.Code == "23502": // not_null_violation
		appErr := AutoSource(
			NewRepoError(ErrRepoConstraint,
				"null value constraint violation"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("constraint", pgErr.ConstraintName).
			WithField("column", pgErr.ColumnName).
			WithField("pg_code", pgErr.Code).
			WithDetails("Required field cannot be null")

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr

	case pgErr.Code == "23514": // check_violation
		appErr := AutoSource(
			NewRepoError(ErrRepoConstraint,
				"check constraint violation"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("constraint", pgErr.ConstraintName).
			WithField("pg_code", pgErr.Code).
			WithDetails(pgErr.Message)

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr

	case strings.HasPrefix(pgErr.Code, "23"): // Other integrity constraints
		appErr := AutoSource(
			NewRepoError(ErrRepoConstraint,
				"database constraint violation"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("constraint", pgErr.ConstraintName).
			WithField("pg_code", pgErr.Code).
			WithDetails(pgErr.Message)

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr

	// ============================================================================
	// Class 25 — Invalid Transaction State
	// ============================================================================
	case strings.HasPrefix(pgErr.Code, "25"):
		return handleTransactionError(pgErr, table, extraFields)

	// ============================================================================
	// Class 28 — Invalid Authorization Specification
	// ============================================================================
	case strings.HasPrefix(pgErr.Code, "28"):
		return handleAuthError(pgErr, table, extraFields)

	// ============================================================================
	// Class 40 — Transaction Rollback
	// ============================================================================
	case strings.HasPrefix(pgErr.Code, "40"):
		return handleTransactionRollback(pgErr, table, extraFields)

	// ============================================================================
	// Class 42 — Syntax Error or Access Rule Violation
	// ============================================================================
	case strings.HasPrefix(pgErr.Code, "42"):
		return handleSyntaxOrAccessError(pgErr, table, extraFields)

	// ============================================================================
	// Class 53 — Insufficient Resources
	// ============================================================================
	case strings.HasPrefix(pgErr.Code, "53"):
		return handleResourceError(pgErr, table, extraFields)

	// ============================================================================
	// Class 54 — Program Limit Exceeded
	// ============================================================================
	case strings.HasPrefix(pgErr.Code, "54"):
		return handleProgramLimitError(pgErr, table, extraFields)

	// ============================================================================
	// Class 55 — Object Not In Prerequisite State
	// ============================================================================
	case strings.HasPrefix(pgErr.Code, "55"):
		return handlePrerequisiteError(pgErr, table, extraFields)

	// ============================================================================
	// Class 57 — Operator Intervention
	// ============================================================================
	case strings.HasPrefix(pgErr.Code, "57"):
		return handleOperatorIntervention(pgErr, table, extraFields)

	// ============================================================================
	// Class 58 — System Error
	// ============================================================================
	case strings.HasPrefix(pgErr.Code, "58"):
		return handleSystemError(pgErr, table, extraFields)

	// ============================================================================
	// Class F0 — Configuration File Error
	// ============================================================================
	case strings.HasPrefix(pgErr.Code, "F0"):
		return handleConfigError(pgErr, table, extraFields)

	// ============================================================================
	// Class HV — Foreign Data Wrapper Error
	// ============================================================================
	case strings.HasPrefix(pgErr.Code, "HV"):
		return handleFDWError(pgErr, table, extraFields)

	// ============================================================================
	// Class P0 — PL/pgSQL Error
	// ============================================================================
	case strings.HasPrefix(pgErr.Code, "P0"):
		return handlePLpgSQLError(pgErr, table, extraFields)

	// ============================================================================
	// Class XX — Internal Error
	// ============================================================================
	case strings.HasPrefix(pgErr.Code, "XX"):
		return handleInternalError(pgErr, table, extraFields)

	// ============================================================================
	// Default: Generic PostgreSQL error
	// ============================================================================
	default:
		appErr := AutoSource(
			WrapRepoError(err, ErrRepoDatabase,
				"database error occurred"))

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
}

// handleConnectionError handles Class 08 errors
func handleConnectionError(pgErr *pgconn.PgError, table string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoConnection,
			"database connection error"))

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

// handleDataException handles Class 22 errors
func handleDataException(pgErr *pgconn.PgError, table string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoDatabase,
			"data validation error"))

	if table != "" {
		_ = appErr.WithField("table", table)
	}
	_ = appErr.WithField("pg_code", pgErr.Code).
		WithField("pg_severity", pgErr.Severity).
		WithField("column", pgErr.ColumnName).
		WithDetails(pgErr.Message)

	for key, value := range extraFields {
		_ = appErr.WithField(key, value)
	}
	return appErr
}

// handleTransactionError handles Class 25 errors
func handleTransactionError(pgErr *pgconn.PgError, table string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoTransaction,
			"invalid transaction state"))

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

// handleTransactionRollback handles Class 40 errors
func handleTransactionRollback(pgErr *pgconn.PgError, table string, extraFields map[string]any) *AppError {
	// 40P01 is deadlock_detected - special case
	if pgErr.Code == "40P01" {
		appErr := AutoSource(
			NewRepoError(ErrRepoDatabase,
				"deadlock detected"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("pg_code", pgErr.Code).
			WithDetails("Transaction deadlock detected, please retry")

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr
	}

	appErr := AutoSource(
		NewRepoError(ErrRepoTransaction,
			"transaction rollback"))

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

// handleResourceError handles Class 53 errors
func handleResourceError(pgErr *pgconn.PgError, table string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoConnection,
			"insufficient database resources"))

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

// handleProgramLimitError handles Class 54 errors
func handleProgramLimitError(pgErr *pgconn.PgError, table string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoDatabase,
			"program limit exceeded"))

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

// handlePrerequisiteError handles Class 55 errors (e.g., lock not available)
func handlePrerequisiteError(pgErr *pgconn.PgError, table string, extraFields map[string]any) *AppError {
	// 55P03 is lock_not_available
	if pgErr.Code == "55P03" {
		appErr := AutoSource(
			NewRepoError(ErrRepoTimeout,
				"lock timeout"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("pg_code", pgErr.Code).
			WithDetails("Could not acquire lock on requested resource")

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr
	}

	appErr := AutoSource(
		NewRepoError(ErrRepoDatabase,
			"object not in prerequisite state"))

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

// handleOperatorIntervention handles Class 57 errors
func handleOperatorIntervention(pgErr *pgconn.PgError, table string, extraFields map[string]any) *AppError {
	// 57014 is query_canceled
	if pgErr.Code == "57014" {
		appErr := AutoSource(
			NewRepoError(ErrRepoTimeout,
				"query canceled"))

		if table != "" {
			_ = appErr.WithField("table", table)
		}
		_ = appErr.WithField("pg_code", pgErr.Code).
			WithDetails("Query was canceled by user or timeout")

		for key, value := range extraFields {
			_ = appErr.WithField(key, value)
		}
		return appErr
	}

	appErr := AutoSource(
		NewRepoError(ErrRepoConnection,
			"database is unavailable"))

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

// handleSystemError handles Class 58 errors
func handleSystemError(pgErr *pgconn.PgError, table string, extraFields map[string]any) *AppError {
	appErr := AutoSource(
		NewRepoError(ErrRepoDatabase,
			"system error"))

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

// ============================================================================
// PostgreSQL Error Codes - Complete Reference
// ============================================================================
// Source: https://www.postgresql.org/docs/current/errcodes-appendix.html
// PostgreSQL v18.1
//
// All messages emitted by the PostgreSQL server are assigned five-character
// error codes that follow the SQL standard's conventions for "SQLSTATE" codes.
//
// The first two characters of an error code denote a class of errors,
// while the last three characters indicate a specific condition within that class.
//
// ============================================================================
// Class 00 — Successful Completion
// ============================================================================
//   00000 - successful_completion
//
// ============================================================================
// Class 01 — Warning
// ============================================================================
//   01000 - warning
//   0100C - dynamic_result_sets_returned
//   01008 - implicit_zero_bit_padding
//   01003 - null_value_eliminated_in_set_function
//   01007 - privilege_not_granted
//   01006 - privilege_not_revoked
//   01004 - string_data_right_truncation
//   01P01 - deprecated_feature
//
// ============================================================================
// Class 02 — No Data
// ============================================================================
//   02000 - no_data
//   02001 - no_additional_dynamic_result_sets_returned
//
// ============================================================================
// Class 03 — SQL Statement Not Yet Complete
// ============================================================================
//   03000 - sql_statement_not_yet_complete
//
// ============================================================================
// Class 08 — Connection Exception
// ============================================================================
//   08000 - connection_exception
//   08003 - connection_does_not_exist
//   08006 - connection_failure
//   08001 - sqlclient_unable_to_establish_sqlconnection
//   08004 - sqlserver_rejected_establishment_of_sqlconnection
//   08007 - transaction_resolution_unknown
//   08P01 - protocol_violation
//
// ============================================================================
// Class 09 — Triggered Action Exception
// ============================================================================
//   09000 - triggered_action_exception
//
// ============================================================================
// Class 0A — Feature Not Supported
// ============================================================================
//   0A000 - feature_not_supported
//
// ============================================================================
// Class 0B — Invalid Transaction Initiation
// ============================================================================
//   0B000 - invalid_transaction_initiation
//
// ============================================================================
// Class 22 — Data Exception
// ============================================================================
//   22000 - data_exception
//   2202E - array_subscript_error
//   22021 - character_not_in_repertoire
//   22008 - datetime_field_overflow
//   22012 - division_by_zero
//   22005 - error_in_assignment
//   2200B - escape_character_conflict
//   22022 - indicator_overflow
//   22015 - interval_field_overflow
//   2201E - invalid_argument_for_logarithm
//   22014 - invalid_argument_for_ntile_function
//   22016 - invalid_argument_for_nth_value_function
//   2201F - invalid_argument_for_power_function
//   2201G - invalid_argument_for_width_bucket_function
//   22018 - invalid_character_value_for_cast
//   22007 - invalid_datetime_format
//   22019 - invalid_escape_character
//   22023 - invalid_parameter_value
//   2201B - invalid_regular_expression
//   2201W - invalid_row_count_in_limit_clause
//   22009 - invalid_time_zone_displacement_value
//   22004 - null_value_not_allowed
//   22002 - null_value_no_indicator_parameter
//   22003 - numeric_value_out_of_range
//   22001 - string_data_right_truncation
//   22011 - substring_error
//   22024 - unterminated_c_string
//   2200F - zero_length_character_string
//   22P01 - floating_point_exception
//   22P02 - invalid_text_representation
//   22P03 - invalid_binary_representation
//   22P04 - bad_copy_file_format
//   22P05 - untranslatable_character
//
// ============================================================================
// Class 23 — Integrity Constraint Violation
// ============================================================================
//   23000 - integrity_constraint_violation
//   23001 - restrict_violation
//   23502 - not_null_violation
//   23503 - foreign_key_violation
//   23505 - unique_violation
//   23514 - check_violation
//   23P01 - exclusion_violation
//
// ============================================================================
// Class 24 — Invalid Cursor State
// ============================================================================
//   24000 - invalid_cursor_state
//
// ============================================================================
// Class 25 — Invalid Transaction State
// ============================================================================
//   25000 - invalid_transaction_state
//   25001 - active_sql_transaction
//   25002 - branch_transaction_already_active
//   25008 - held_cursor_requires_same_isolation_level
//   25003 - inappropriate_access_mode_for_branch_transaction
//   25004 - inappropriate_isolation_level_for_branch_transaction
//   25005 - no_active_sql_transaction_for_branch_transaction
//   25006 - read_only_sql_transaction
//   25007 - schema_and_data_statement_mixing_not_supported
//   25P01 - no_active_sql_transaction
//   25P02 - in_failed_sql_transaction
//   25P03 - idle_in_transaction_session_timeout
//   25P04 - transaction_timeout
//
// ============================================================================
// Class 26 — Invalid SQL Statement Name
// ============================================================================
//   26000 - invalid_sql_statement_name
//
// ============================================================================
// Class 27 — Triggered Data Change Violation
// ============================================================================
//   27000 - triggered_data_change_violation
//
// ============================================================================
// Class 28 — Invalid Authorization Specification
// ============================================================================
//   28000 - invalid_authorization_specification
//   28P01 - invalid_password
//
// ============================================================================
// Class 2B — Dependent Privilege Descriptors Still Exist
// ============================================================================
//   2B000 - dependent_privilege_descriptors_still_exist
//   2BP01 - dependent_objects_still_exist
//
// ============================================================================
// Class 2D — Invalid Transaction Termination
// ============================================================================
//   2D000 - invalid_transaction_termination
//
// ============================================================================
// Class 2F — SQL Routine Exception
// ============================================================================
//   2F000 - sql_routine_exception
//   2F005 - function_executed_no_return_statement
//   2F002 - modifying_sql_data_not_permitted
//   2F003 - prohibited_sql_statement_attempted
//   2F004 - reading_sql_data_not_permitted
//
// ============================================================================
// Class 40 — Transaction Rollback
// ============================================================================
//   40000 - transaction_rollback
//   40002 - transaction_integrity_constraint_violation
//   40001 - serialization_failure
//   40003 - statement_completion_unknown
//   40P01 - deadlock_detected
//
// ============================================================================
// Class 42 — Syntax Error or Access Rule Violation
// ============================================================================
//   42000 - syntax_error_or_access_rule_violation
//   42601 - syntax_error
//   42501 - insufficient_privilege
//   42846 - cannot_coerce
//   42803 - grouping_error
//   42P20 - windowing_error
//   42P19 - invalid_recursion
//   42830 - invalid_foreign_key
//   42602 - invalid_name
//   42622 - name_too_long
//   42939 - reserved_name
//   42804 - datatype_mismatch
//   42P18 - indeterminate_datatype
//   42P21 - collation_mismatch
//   42P22 - indeterminate_collation
//   42809 - wrong_object_type
//   428C9 - generated_always
//   42703 - undefined_column
//   42883 - undefined_function
//   42P01 - undefined_table
//   42P02 - undefined_parameter
//   42704 - undefined_object
//   42701 - duplicate_column
//   42P03 - duplicate_cursor
//   42P04 - duplicate_database
//   42723 - duplicate_function
//   42P05 - duplicate_prepared_statement
//   42P06 - duplicate_schema
//   42P07 - duplicate_table
//   42712 - duplicate_alias
//   42710 - duplicate_object
//   42702 - ambiguous_column
//   42725 - ambiguous_function
//   42P08 - ambiguous_parameter
//   42P09 - ambiguous_alias
//   42P10 - invalid_column_reference
//   42611 - invalid_column_definition
//   42P11 - invalid_cursor_definition
//   42P12 - invalid_database_definition
//   42P13 - invalid_function_definition
//   42P14 - invalid_prepared_statement_definition
//   42P15 - invalid_schema_definition
//   42P16 - invalid_table_definition
//   42P17 - invalid_object_definition
//
// ============================================================================
// Class 44 — WITH CHECK OPTION Violation
// ============================================================================
//   44000 - with_check_option_violation
//
// ============================================================================
// Class 53 — Insufficient Resources
// ============================================================================
//   53000 - insufficient_resources
//   53100 - disk_full
//   53200 - out_of_memory
//   53300 - too_many_connections
//   53400 - configuration_limit_exceeded
//
// ============================================================================
// Class 54 — Program Limit Exceeded
// ============================================================================
//   54000 - program_limit_exceeded
//   54001 - statement_too_complex
//   54011 - too_many_columns
//   54023 - too_many_arguments
//
// ============================================================================
// Class 55 — Object Not In Prerequisite State
// ============================================================================
//   55000 - object_not_in_prerequisite_state
//   55006 - object_in_use
//   55P02 - cant_change_runtime_param
//   55P03 - lock_not_available
//   55P04 - unsafe_new_enum_value_usage
//
// ============================================================================
// Class 57 — Operator Intervention
// ============================================================================
//   57000 - operator_intervention
//   57014 - query_canceled
//   57P01 - admin_shutdown
//   57P02 - crash_shutdown
//   57P03 - cannot_connect_now
//   57P04 - database_dropped
//   57P05 - idle_session_timeout
//
// ============================================================================
// Class 58 — System Error (errors external to PostgreSQL)
// ============================================================================
//   58000 - system_error
//   58030 - io_error
//   58P01 - undefined_file
//   58P02 - duplicate_file
//   58P03 - file_name_too_long
//
// ============================================================================
// Class F0 — Configuration File Error
// ============================================================================
//   F0000 - config_file_error
//   F0001 - lock_file_exists
//
// ============================================================================
// Class HV — Foreign Data Wrapper Error (SQL/MED)
// ============================================================================
//   HV000 - fdw_error
//   HV005 - fdw_column_name_not_found
//   HV002 - fdw_dynamic_parameter_value_needed
//   HV010 - fdw_function_sequence_error
//   HV021 - fdw_inconsistent_descriptor_information
//   HV024 - fdw_invalid_attribute_value
//   HV007 - fdw_invalid_column_name
//   HV008 - fdw_invalid_column_number
//   HV004 - fdw_invalid_data_type
//   HV006 - fdw_invalid_data_type_descriptors
//   HV091 - fdw_invalid_descriptor_field_identifier
//   HV00B - fdw_invalid_handle
//   HV00C - fdw_invalid_option_index
//   HV00D - fdw_invalid_option_name
//   HV090 - fdw_invalid_string_length_or_buffer_length
//   HV00A - fdw_invalid_string_format
//   HV009 - fdw_invalid_use_of_null_pointer
//   HV014 - fdw_too_many_handles
//   HV001 - fdw_out_of_memory
//   HV00P - fdw_no_schemas
//   HV00J - fdw_option_name_not_found
//   HV00K - fdw_reply_handle
//   HV00Q - fdw_schema_not_found
//   HV00R - fdw_table_not_found
//   HV00L - fdw_unable_to_create_execution
//   HV00M - fdw_unable_to_create_reply
//   HV00N - fdw_unable_to_establish_connection
//
// ============================================================================
// Class P0 — PL/pgSQL Error
// ============================================================================
//   P0000 - plpgsql_error
//   P0001 - raise_exception
//   P0002 - no_data_found
//   P0003 - too_many_rows
//   P0004 - assert_failure
//
// ============================================================================
// Class XX — Internal Error
// ============================================================================
//   XX000 - internal_error
//   XX001 - data_corrupted
//   XX002 - index_corrupted
