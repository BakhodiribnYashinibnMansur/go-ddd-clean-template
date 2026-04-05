package errorx

import "testing"

func TestPgErrCodeConstants(t *testing.T) {
	// Verify some key constants have expected values
	tests := []struct {
		name  string
		code  string
		value string
	}{
		{"SuccessfulCompletion", SuccessfulCompletion, "00000"},
		{"UniqueViolation", UniqueViolation, "23505"},
		{"ForeignKeyViolation", ForeignKeyViolation, "23503"},
		{"NotNullViolation", NotNullViolation, "23502"},
		{"CheckViolation", CheckViolation, "23514"},
		{"ConnectionFailure", ConnectionFailure, "08006"},
		{"DeadlockDetected", DeadlockDetected, "40P01"},
		{"SerializationFailure", SerializationFailure, "40001"},
		{"SyntaxError", SyntaxError, "42601"},
		{"InsufficientPrivilege", InsufficientPrivilege, "42501"},
		{"UndefinedTable", UndefinedTable, "42P01"},
		{"QueryCanceled", QueryCanceled, "57014"},
		{"DiskFull", DiskFull, "53100"},
		{"OutOfMemory", OutOfMemory, "53200"},
		{"TooManyConnections", TooManyConnections, "53300"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.value {
				t.Errorf("expected %q, got %q", tt.value, tt.code)
			}
		})
	}
}

func TestIsIntegrityConstraintViolation(t *testing.T) {
	tests := []struct {
		name string
		code string
		want bool
	}{
		{"unique violation", UniqueViolation, true},
		{"foreign key violation", ForeignKeyViolation, true},
		{"not null violation", NotNullViolation, true},
		{"check violation", CheckViolation, true},
		{"connection failure", ConnectionFailure, false},
		{"syntax error", SyntaxError, false},
		{"empty string", "", false},
		{"short string", "2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsIntegrityConstraintViolation(tt.code)
			if got != tt.want {
				t.Errorf("IsIntegrityConstraintViolation(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestIsConnectionException(t *testing.T) {
	tests := []struct {
		name string
		code string
		want bool
	}{
		{"connection exception", ConnectionException, true},
		{"connection failure", ConnectionFailure, true},
		{"connection does not exist", ConnectionDoesNotExist, true},
		{"unique violation", UniqueViolation, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsConnectionException(tt.code)
			if got != tt.want {
				t.Errorf("IsConnectionException(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestIsDataException(t *testing.T) {
	tests := []struct {
		name string
		code string
		want bool
	}{
		{"data exception", DataException, true},
		{"division by zero", DivisionByZero, true},
		{"invalid text representation", InvalidTextRepresentation, true},
		{"connection failure", ConnectionFailure, false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDataException(tt.code)
			if got != tt.want {
				t.Errorf("IsDataException(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestIsInvalidTransactionState(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{InvalidTransactionState, true},
		{NoActiveSQLTransaction, true},
		{ConnectionFailure, false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if got := IsInvalidTransactionState(tt.code); got != tt.want {
				t.Errorf("IsInvalidTransactionState(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestIsTransactionRollback(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{TransactionRollback, true},
		{DeadlockDetected, true},
		{SerializationFailure, true},
		{ConnectionFailure, false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if got := IsTransactionRollback(tt.code); got != tt.want {
				t.Errorf("IsTransactionRollback(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestIsSyntaxErrorOrAccessRuleViolation(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{SyntaxError, true},
		{InsufficientPrivilege, true},
		{UndefinedTable, true},
		{ConnectionFailure, false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if got := IsSyntaxErrorOrAccessRuleViolation(tt.code); got != tt.want {
				t.Errorf("IsSyntaxErrorOrAccessRuleViolation(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestIsInsufficientResources(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{InsufficientResources, true},
		{DiskFull, true},
		{OutOfMemory, true},
		{TooManyConnections, true},
		{ConnectionFailure, false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if got := IsInsufficientResources(tt.code); got != tt.want {
				t.Errorf("IsInsufficientResources(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestIsProgramLimitExceeded(t *testing.T) {
	if !IsProgramLimitExceeded(ProgramLimitExceeded) {
		t.Error("expected true for ProgramLimitExceeded")
	}
	if IsProgramLimitExceeded(ConnectionFailure) {
		t.Error("expected false for ConnectionFailure")
	}
}

func TestIsObjectNotInPrerequisiteState(t *testing.T) {
	if !IsObjectNotInPrerequisiteState(ObjectNotInPrerequisiteState) {
		t.Error("expected true for ObjectNotInPrerequisiteState")
	}
	if !IsObjectNotInPrerequisiteState(LockNotAvailable) {
		t.Error("expected true for LockNotAvailable")
	}
}

func TestIsOperatorIntervention(t *testing.T) {
	if !IsOperatorIntervention(QueryCanceled) {
		t.Error("expected true for QueryCanceled")
	}
	if !IsOperatorIntervention(AdminShutdown) {
		t.Error("expected true for AdminShutdown")
	}
	if IsOperatorIntervention(ConnectionFailure) {
		t.Error("expected false for ConnectionFailure")
	}
}

func TestIsSystemError(t *testing.T) {
	if !IsSystemError(SystemError) {
		t.Error("expected true for SystemError")
	}
	if !IsSystemError(IOError) {
		t.Error("expected true for IOError")
	}
	if IsSystemError(ConnectionFailure) {
		t.Error("expected false for ConnectionFailure")
	}
}

func TestIsForeignDataWrapperError(t *testing.T) {
	if !IsForeignDataWrapperError(FDWError) {
		t.Error("expected true for FDWError")
	}
	if IsForeignDataWrapperError(ConnectionFailure) {
		t.Error("expected false for ConnectionFailure")
	}
}

func TestIsPLpgSQLError(t *testing.T) {
	if !IsPLpgSQLError(PLpgSQLError) {
		t.Error("expected true for PLpgSQLError")
	}
	if !IsPLpgSQLError(RaiseException) {
		t.Error("expected true for RaiseException")
	}
	if IsPLpgSQLError(ConnectionFailure) {
		t.Error("expected false for ConnectionFailure")
	}
}

func TestIsInternalError(t *testing.T) {
	if !IsInternalError(InternalError) {
		t.Error("expected true for InternalError")
	}
	if !IsInternalError(DataCorrupted) {
		t.Error("expected true for DataCorrupted")
	}
	if IsInternalError(ConnectionFailure) {
		t.Error("expected false for ConnectionFailure")
	}
}
