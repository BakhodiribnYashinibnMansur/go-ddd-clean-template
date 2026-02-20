package systemerror_test

import (
	"context"
	"testing"

	"gct/internal/domain"
	systemerror "gct/internal/repo/persistent/postgres/systemerror"
	"gct/pkg/db/postgres"
	"gct/pkg/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This is an example test showing how to use the system error repository
// Note: This requires a running database for integration testing

func TestSystemError_Create(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup (you would initialize your actual DB connection here)
	// pg := postgres.New(...)
	// log := logger.New(...)
	// repo := systemerror.New(pg, log)

	t.Run("create error with minimal info", func(t *testing.T) {
		t.Skip("Integration test - requires DB")

		var repo *systemerror.Repo // Initialize with actual repo
		ctx := context.Background()

		input := systemerror.CreateSystemErrorInput{
			Code:     "TEST_ERROR",
			Message:  "This is a test error",
			Severity: "ERROR",
		}

		result, err := repo.CreateSystemError(ctx, input)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "TEST_ERROR", result.Code)
		assert.Equal(t, "ERROR", result.Severity)
	})

	t.Run("create error with full context", func(t *testing.T) {
		t.Skip("Integration test - requires DB")

		var repo *systemerror.Repo // Initialize with actual repo
		ctx := context.Background()

		requestID := uuid.New()
		userID := uuid.New()
		ipAddr := "192.168.1.1"
		path := "/api/v1/users"
		method := "POST"
		stackTrace := "stack trace here"

		input := systemerror.CreateSystemErrorInput{
			Code:        "AUTH_FAILED",
			Message:     "Authentication failed",
			StackTrace:  &stackTrace,
			Severity:    "ERROR",
			ServiceName: "auth-service",
			RequestID:   &requestID,
			UserID:      &userID,
			IPAddress:   &ipAddr,
			Path:        &path,
			Method:      &method,
			Metadata: map[string]any{
				"attempt": 3,
				"reason":  "invalid_password",
			},
		}

		result, err := repo.CreateSystemError(ctx, input)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "AUTH_FAILED", result.Code)
		assert.Equal(t, requestID, *result.RequestID)
		assert.Equal(t, userID, *result.UserID)
	})
}

func TestSystemError_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	t.Run("get existing error", func(t *testing.T) {
		t.Skip("Integration test - requires DB")

		var repo *systemerror.Repo // Initialize with actual repo
		ctx := context.Background()

		// First create an error
		input := systemerror.CreateSystemErrorInput{
			Code:     "TEST_ERROR",
			Message:  "Test error",
			Severity: "ERROR",
		}

		created, err := repo.CreateSystemError(ctx, input)
		require.NoError(t, err)

		// Then retrieve it
		result, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, result.ID)
		assert.Equal(t, created.Code, result.Code)
	})
}

func TestSystemError_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	t.Run("list with filters", func(t *testing.T) {
		t.Skip("Integration test - requires DB")

		var repo *systemerror.Repo // Initialize with actual repo
		ctx := context.Background()

		// Create some test errors
		codes := []string{"ERROR_1", "ERROR_2", "ERROR_3"}
		for _, code := range codes {
			_, err := repo.CreateSystemError(ctx, systemerror.CreateSystemErrorInput{
				Code:     code,
				Message:  "Test error: " + code,
				Severity: "ERROR",
			})
			require.NoError(t, err)
		}

		// List all errors
		filter := systemerror.ListFilter{
			Pagination: &domain.Pagination{Limit: 10},
		}

		results, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 3)
	})

	t.Run("list with code filter", func(t *testing.T) {
		t.Skip("Integration test - requires DB")

		var repo *systemerror.Repo // Initialize with actual repo
		ctx := context.Background()

		code := "SPECIFIC_ERROR"
		filter := systemerror.ListFilter{
			SystemErrorFilter: domain.SystemErrorFilter{Code: &code},
			Pagination:        &domain.Pagination{Limit: 10},
		}

		results, err := repo.List(ctx, filter)
		require.NoError(t, err)
		for _, result := range results {
			assert.Equal(t, code, result.Code)
		}
	})
}

func TestSystemError_MarkAsResolved(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	t.Run("mark error as resolved", func(t *testing.T) {
		t.Skip("Integration test - requires DB")

		var repo *systemerror.Repo // Initialize with actual repo
		ctx := context.Background()

		// Create an error
		input := systemerror.CreateSystemErrorInput{
			Code:     "RESOLVABLE_ERROR",
			Message:  "This error will be resolved",
			Severity: "ERROR",
		}

		created, err := repo.CreateSystemError(ctx, input)
		require.NoError(t, err)
		assert.False(t, created.IsResolved)

		// Mark as resolved
		resolverID := uuid.New()
		err = repo.MarkAsResolved(ctx, created.ID, resolverID)
		require.NoError(t, err)

		// Verify it's resolved
		result, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		assert.True(t, result.IsResolved)
		assert.NotNil(t, result.ResolvedAt)
		assert.Equal(t, resolverID, *result.ResolvedBy)
	})
}

// Example of how to use the error logger in your code
func Example() {
	// This is not a real test, just an example
	var pg *postgres.Postgres
	var log logger.Log

	// Initialize repository
	repo := systemerror.New(pg, log)
	ctx := context.Background()

	// Example 1: Log a simple error
	_, _ = repo.CreateSystemError(ctx, systemerror.CreateSystemErrorInput{
		Code:     "USER_NOT_FOUND",
		Message:  "User with ID 123 not found",
		Severity: "ERROR",
	})

	// Example 2: Log error with full context
	requestID := uuid.New()
	userID := uuid.New()
	ipAddr := "192.168.1.1"
	path := "/api/v1/users/123"
	method := "GET"

	_, _ = repo.CreateSystemError(ctx, systemerror.CreateSystemErrorInput{
		Code:        "DATABASE_ERROR",
		Message:     "Failed to query user table",
		Severity:    "ERROR",
		ServiceName: "user-service",
		RequestID:   &requestID,
		UserID:      &userID,
		IPAddress:   &ipAddr,
		Path:        &path,
		Method:      &method,
		Metadata: map[string]any{
			"table":     "users",
			"operation": "SELECT",
			"duration":  "1.5s",
		},
	})

	// Example 3: Query errors
	code := "DATABASE_ERROR"
	errors, _ := repo.List(ctx, systemerror.ListFilter{
		SystemErrorFilter: domain.SystemErrorFilter{Code: &code},
		Pagination:        &domain.Pagination{Limit: 100},
	})

	// Example 4: Mark error as resolved
	if len(errors) > 0 {
		resolverID := uuid.New()
		_ = repo.MarkAsResolved(ctx, errors[0].ID, resolverID)
	}
}
