package main

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/pkg/errors"
	"go.uber.org/zap"
)

// This example demonstrates the CORRECT way to handle errors:
// - Repository: Return error with context (NO LOGGING)
// - Service: Map error and add context (NO LOGGING)
// - Handler: LOG ONCE with full trace from all layers

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	ctx := context.Background()

	fmt.Println("=== Correct Logging Pattern ===")
	fmt.Println("Only Handler logs. Lower layers return errors with context.")

	// Simulate the full flow
	simulateRequestFlow(ctx, logger, "user123")
}

func simulateRequestFlow(ctx context.Context, logger *zap.Logger, userID string) {
	fmt.Println("1. Request arrives at Handler")
	fmt.Println("2. Handler calls Service")
	fmt.Println("3. Service calls Repository")
	fmt.Println("4. Repository returns error (NO LOG)")
	fmt.Println("5. Service maps error (NO LOG)")
	fmt.Println("6. Handler logs ONCE with full trace")

	// === REPOSITORY LAYER ===
	fmt.Println("--- Repository Layer (Silent) ---")
	repoErr := simulateRepository(ctx, userID)
	fmt.Printf("✗ Error returned: %s\n", repoErr.Type)
	fmt.Printf("  From: %v (function: %v)\n\n",
		repoErr.Fields["file"], repoErr.Fields["function"])

	// === SERVICE LAYER ===
	fmt.Println("--- Service Layer (Silent) ---")
	serviceErr := simulateService(ctx, repoErr)
	fmt.Printf("✗ Error mapped: %s\n", serviceErr.Type)
	fmt.Printf("  From: %v (function: %v)\n\n",
		serviceErr.Fields["file"], serviceErr.Fields["function"])

	// === HANDLER LAYER ===
	fmt.Println("--- Handler Layer (LOGS HERE) ---")
	handlerErr := simulateHandler(ctx, serviceErr)
	fmt.Printf("✓ Logging error: %s\n", handlerErr.Type)
	fmt.Printf("  From: %v (function: %v)\n\n",
		handlerErr.Fields["file"], handlerErr.Fields["function"])

	// LOG ONCE - with all layer information
	fmt.Println("=== SINGLE LOG OUTPUT ===")
	errors.LogError(logger, handlerErr)

	fmt.Println("\n=== Log Contains ===")
	fmt.Println("✓ Repository file:", handlerErr.Fields["repo_file"])
	fmt.Println("✓ Service file:", handlerErr.Fields["service_file"])
	fmt.Println("✓ Handler file:", handlerErr.Fields["file"])
	fmt.Println("✓ Complete error trace through all layers")
}

// ============================================================================
// REPOSITORY LAYER - Returns error, NO LOGGING
// ============================================================================

func simulateRepository(ctx context.Context, userID string) *errors.AppError {
	// Simulate database error
	err := errors.NewRepoError(ctx, errors.ErrRepoNotFound, "user not found in database").
		WithField("user_id", userID).
		WithField("table", "users").
		WithField("file", "internal/repo/persistent/postgres/user.go").
		WithField("function", "GetByID").
		WithField("line", "42").
		WithDetails("No user record exists with the given ID in database")

	// NO LOGGING HERE!
	// Just return error with context
	return err
}

// ============================================================================
// SERVICE LAYER - Maps error, NO LOGGING
// ============================================================================

func simulateService(ctx context.Context, repoErr *errors.AppError) *errors.AppError {
	// Map repository error to service error
	serviceErr := errors.MapRepoToServiceError(ctx, repoErr)

	// Add service layer context
	serviceErr.
		WithField("service_file", "internal/usecase/user/service.go").
		WithField("service_function", "GetUser").
		WithField("service_line", "28").
		WithField("operation", "get_user")

	// Preserve repository context
	if repoFile, ok := repoErr.Fields["file"]; ok {
		serviceErr.WithField("repo_file", repoFile)
	}
	if repoFunc, ok := repoErr.Fields["function"]; ok {
		serviceErr.WithField("repo_function", repoFunc)
	}

	// NO LOGGING HERE!
	// Just return mapped error
	return serviceErr
}

// ============================================================================
// HANDLER LAYER - Maps error and LOGS ONCE
// ============================================================================

func simulateHandler(ctx context.Context, serviceErr *errors.AppError) *errors.AppError {
	// Map service error to handler error
	handlerErr := errors.MapServiceToHandlerError(ctx, serviceErr)

	// Add handler layer context
	handlerErr.
		WithField("file", "internal/controller/restapi/user/handler.go").
		WithField("function", "GetUser").
		WithField("line", "67").
		WithField("endpoint", "/api/v1/users/123").
		WithField("method", "GET").
		WithField("request_id", "req-abc-123")

	// Preserve service context
	if svcFile, ok := serviceErr.Fields["service_file"]; ok {
		handlerErr.WithField("service_file", svcFile)
	}
	if svcFunc, ok := serviceErr.Fields["service_function"]; ok {
		handlerErr.WithField("service_function", svcFunc)
	}

	// Preserve repository context
	if repoFile, ok := serviceErr.Fields["repo_file"]; ok {
		handlerErr.WithField("repo_file", repoFile)
	}
	if repoFunc, ok := serviceErr.Fields["repo_function"]; ok {
		handlerErr.WithField("repo_function", repoFunc)
	}

	// LOG ONCE HERE!
	// This will be the ONLY log entry for this error
	return handlerErr
}
