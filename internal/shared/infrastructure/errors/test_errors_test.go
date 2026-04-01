package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestTestErrors_AreDistinct(t *testing.T) {
	// Verify all test errors are distinct from each other
	testErrors := []struct {
		name string
		err  error
	}{
		{"ErrTestDB", ErrTestDB},
		{"ErrTestDBConnection", ErrTestDBConnection},
		{"ErrTestConnectionTimeout", ErrTestConnectionTimeout},
		{"ErrTestDuplicateKey", ErrTestDuplicateKey},
		{"ErrTestForeignKey", ErrTestForeignKey},
		{"ErrTestInvalidPhone", ErrTestInvalidPhone},
		{"ErrTestUserNotFound", ErrTestUserNotFound},
		{"ErrTestSessionNotFound", ErrTestSessionNotFound},
		{"ErrTestHandler", ErrTestHandler},
		{"ErrTestService", ErrTestService},
		{"ErrTestRegular", ErrTestRegular},
		{"ErrTestBase", ErrTestBase},
	}

	for i, a := range testErrors {
		for j, b := range testErrors {
			if i != j {
				if errors.Is(a.err, b.err) {
					t.Errorf("expected %s and %s to be distinct errors", a.name, b.name)
				}
			}
		}
	}
}

func TestTestErrors_CanBeWrapped(t *testing.T) {
	wrapped := fmt.Errorf("context info: %w", ErrTestDB)

	if !errors.Is(wrapped, ErrTestDB) {
		t.Error("expected wrapped error to match ErrTestDB")
	}

	if errors.Is(wrapped, ErrTestDBConnection) {
		t.Error("expected wrapped error to NOT match ErrTestDBConnection")
	}
}

func TestTestErrors_NotNil(t *testing.T) {
	testErrors := []struct {
		name string
		err  error
	}{
		{"ErrTestDB", ErrTestDB},
		{"ErrTestDBConnection", ErrTestDBConnection},
		{"ErrTestConnectionTimeout", ErrTestConnectionTimeout},
		{"ErrTestSelectQuery", ErrTestSelectQuery},
		{"ErrTestCountError", ErrTestCountError},
		{"ErrTestDuplicateKey", ErrTestDuplicateKey},
		{"ErrTestForeignKey", ErrTestForeignKey},
		{"ErrTestUniqueConstraint", ErrTestUniqueConstraint},
		{"ErrTestCheckConstraint", ErrTestCheckConstraint},
		{"ErrTestInvalidPhone", ErrTestInvalidPhone},
		{"ErrTestInvalidUserID", ErrTestInvalidUserID},
		{"ErrTestInvalidCompanyID", ErrTestInvalidCompanyID},
		{"ErrTestInvalidFilter", ErrTestInvalidFilter},
		{"ErrTestInvalidSession", ErrTestInvalidSession},
		{"ErrTestValidationFailed", ErrTestValidationFailed},
		{"ErrTestPasswordHashEmpty", ErrTestPasswordHashEmpty},
		{"ErrTestPermissionDenied", ErrTestPermissionDenied},
		{"ErrTestDatabaseLocked", ErrTestDatabaseLocked},
		{"ErrTestRedisConnection", ErrTestRedisConnection},
		{"ErrTestRedisScan", ErrTestRedisScan},
		{"ErrTestRedisDialTCP", ErrTestRedisDialTCP},
		{"ErrTestRedisIOTimeout", ErrTestRedisIOTimeout},
		{"ErrTestRedisContextDeadline", ErrTestRedisContextDeadline},
		{"ErrTestRedisWrongPass", ErrTestRedisWrongPass},
		{"ErrTestRedisWrongType", ErrTestRedisWrongType},
		{"ErrTestRedisOOM", ErrTestRedisOOM},
		{"ErrTestRedisReadOnly", ErrTestRedisReadOnly},
		{"ErrTestRedisClusterDown", ErrTestRedisClusterDown},
		{"ErrTestRedisNoScript", ErrTestRedisNoScript},
		{"ErrTestRedisNoAuth", ErrTestRedisNoAuth},
		{"ErrTestRedisEOF", ErrTestRedisEOF},
		{"ErrTestRedisBrokenPipe", ErrTestRedisBrokenPipe},
		{"ErrTestRedisRandom", ErrTestRedisRandom},
		{"ErrTestRedisError", ErrTestRedisError},
		{"ErrTestUserNotFound", ErrTestUserNotFound},
		{"ErrTestSessionNotFound", ErrTestSessionNotFound},
		{"ErrTestSessionCreationFailed", ErrTestSessionCreationFailed},
		{"ErrTestSessionRevokeFailed", ErrTestSessionRevokeFailed},
		{"ErrTestDeleteFailed", ErrTestDeleteFailed},
		{"ErrTestUpdateFailed", ErrTestUpdateFailed},
		{"ErrTestScanError", ErrTestScanError},
		{"ErrTestHandler", ErrTestHandler},
		{"ErrTestService", ErrTestService},
		{"ErrTestRegular", ErrTestRegular},
		{"ErrTestBase", ErrTestBase},
	}

	for _, tt := range testErrors {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s should not be nil", tt.name)
			}
			if tt.err.Error() == "" {
				t.Errorf("%s should have non-empty message", tt.name)
			}
		})
	}
}

func TestTestErrors_ErrorMessages(t *testing.T) {
	if ErrTestDB.Error() != "database error" {
		t.Errorf("ErrTestDB message = %q, want 'database error'", ErrTestDB.Error())
	}
	if ErrTestUserNotFound.Error() != "user not found" {
		t.Errorf("ErrTestUserNotFound message = %q, want 'user not found'", ErrTestUserNotFound.Error())
	}
	if ErrTestDuplicateKey.Error() != "duplicate key value violates unique constraint" {
		t.Errorf("ErrTestDuplicateKey message = %q", ErrTestDuplicateKey.Error())
	}
}

func TestTestErrors_DoubleWrapping(t *testing.T) {
	inner := fmt.Errorf("inner: %w", ErrTestDB)
	outer := fmt.Errorf("outer: %w", inner)

	if !errors.Is(outer, ErrTestDB) {
		t.Error("expected double-wrapped error to match ErrTestDB")
	}
}
