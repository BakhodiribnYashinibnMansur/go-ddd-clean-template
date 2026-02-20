package errorx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// Example_usage shows how to use the error logger
func ExampleUsage() {
	// This is not a real test, just an example showing usage

	// Initialize error logger (in real code, you would pass actual repo and logger)
	// errorLogger := errorx.NewErrorLogger(repo.Persistent.Postgres.SystemError, logger)

	ctx := context.Background()
	_ = ctx // Used in commented examples

	// Example 1: Simple error logging
	// err := errorLogger.LogErrorSimple(ctx, "USER_NOT_FOUND", "User not found", errors.New("no user with id 123"))

	// Example 2: Full error logging with context
	requestID := uuid.New()
	userID := uuid.New()
	ipAddr := "192.168.1.1"
	path := "/api/v1/auth/login"
	method := "POST"

	_ = requestID
	_ = userID
	_ = ipAddr
	_ = path
	_ = method

	/*
		err = errorLogger.LogError(ctx, errorx.LogErrorInput{
			Code:        "AUTH_FAILED",
			Message:     "Authentication failed",
			Err:         errors.New("invalid credentials"),
			Severity:    "ERROR",
			ServiceName: "auth-service",
			RequestID:   &requestID,
			UserID:      &userID,
			IPAddress:   &ipAddr,
			Path:        &path,
			Method:      &method,
			Metadata: map[string]any{
				"attempt":  3,
				"reason":   "invalid_password",
				"username": "john_doe",
			},
		})
	*/

	// Example 3: Different severity levels
	// errorLogger.LogWarn(ctx, "SLOW_QUERY", "Database query took too long", nil)
	// errorLogger.LogFatal(ctx, "DB_CONNECTION_LOST", "Lost connection to database", err)
	// errorLogger.LogPanic(ctx, "SYSTEM_PANIC", "System panic occurred", err)
}

// ShowHandlerUsage demonstrates usage in a real handler
func ShowHandlerUsage() {
	// Pseudo-code showing how to use in a real HTTP handler
	/*
		func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get request context
			requestID := middleware.GetRequestID(ctx)
			ipAddr := httpx.GetClientIP(r)

			// Try to authenticate
			user, err := h.useCase.Authenticate(ctx, username, password)
			if err != nil {
				// Log the error to database
				h.errorLogger.LogError(ctx, errorx.LogErrorInput{
					Code:        "AUTH_FAILED",
					Message:     "Failed to authenticate user",
					Err:         err,
					Severity:    "ERROR",
					ServiceName: "api",
					RequestID:   &requestID,
					IPAddress:   &ipAddr,
					Path:        stringPtr(r.URL.Path),
					Method:      stringPtr(r.Method),
					Metadata: map[string]any{
						"username": username,
						"ip":       ipAddr,
					},
				})

				// Return error response
				httpx.Error(w, http.StatusUnauthorized, "Invalid credentials")
				return
			}

			// Success...
		}
	*/
}

// ShowUseCaseUsage demonstrates usage in a use case
func ShowUseCaseUsage() {
	// Pseudo-code showing how to use in a use case
	/*
		func (uc *UseCase) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
			// Try to create user
			user, err := uc.repo.User.Create(ctx, input)
			if err != nil {
				// Log the error
				uc.errorLogger.LogError(ctx, errorx.LogErrorInput{
					Code:        "USER_CREATE_FAILED",
					Message:     "Failed to create user in database",
					Err:         err,
					Severity:    "ERROR",
					ServiceName: "user-service",
					Metadata: map[string]any{
						"username": input.Username,
						"email":    input.Email,
					},
				})

				return nil, fmt.Errorf("failed to create user: %w", err)
			}

			return user, nil
		}
	*/
}

// ShowErrorCodesOrganization demonstrates error codes organization
func ShowErrorCodesOrganization() {
	// You can organize error codes in constants
	/*
		const (
			// Authentication errors
			ErrCodeAuthFailed          = "AUTH_FAILED"
			ErrCodeInvalidToken        = "INVALID_TOKEN"
			ErrCodeTokenExpired        = "TOKEN_EXPIRED"
			ErrCodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"

			// User errors
			ErrCodeUserNotFound        = "USER_NOT_FOUND"
			ErrCodeUserAlreadyExists   = "USER_ALREADY_EXISTS"
			ErrCodeUserNotApproved     = "USER_NOT_APPROVED"
			ErrCodeUserBlocked         = "USER_BLOCKED"

			// Database errors
			ErrCodeDatabaseError       = "DATABASE_ERROR"
			ErrCodeQueryTimeout        = "QUERY_TIMEOUT"
			ErrCodeConnectionLost      = "CONNECTION_LOST"

			// Validation errors
			ErrCodeValidationFailed    = "VALIDATION_FAILED"
			ErrCodeInvalidInput        = "INVALID_INPUT"
			ErrCodeMissingField        = "MISSING_FIELD"

			// External service errors
			ErrCodeExternalServiceError = "EXTERNAL_SERVICE_ERROR"
			ErrCodeAPITimeout          = "API_TIMEOUT"
			ErrCodeAPIRateLimited      = "API_RATE_LIMITED"
		)
	*/
}

func TestErrorLogger_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	t.Run("log simple error", func(t *testing.T) {
		t.Skip("Integration test - requires DB and logger")

		// var errorLogger *errorx.ErrorLogger
		ctx := context.Background()

		// err := errorLogger.LogErrorSimple(ctx, "TEST_ERROR", "This is a test", errors.New("test error"))
		// assert.NoError(t, err)
		_ = ctx
	})

	t.Run("log error with full context", func(t *testing.T) {
		t.Skip("Integration test - requires DB and logger")

		// var errorLogger *errorx.ErrorLogger
		ctx := context.Background()
		requestID := uuid.New()

		// err := errorLogger.LogError(ctx, errorx.LogErrorInput{
		// 	Code:      "TEST_ERROR",
		// 	Message:   "Test error with context",
		// 	Err:       errors.New("test error"),
		// 	RequestID: &requestID,
		// })
		// assert.NoError(t, err)
		_ = ctx
		_ = requestID
	})

	t.Run("log different severity levels", func(t *testing.T) {
		t.Skip("Integration test - requires DB and logger")

		// var errorLogger *errorx.ErrorLogger
		ctx := context.Background()
		testErr := errors.New("test error")

		// Test all severity levels
		// errorLogger.LogWarn(ctx, "WARN_TEST", "Warning message", nil)
		// errorLogger.LogErrorSimple(ctx, "ERROR_TEST", "Error message", testErr)
		// errorLogger.LogFatal(ctx, "FATAL_TEST", "Fatal message", testErr)
		// errorLogger.LogPanic(ctx, "PANIC_TEST", "Panic message", testErr)

		_ = ctx
		_ = testErr
	})
}
