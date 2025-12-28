package main

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/pkg/errors"
)

// Example: Complete layer-based error handling flow

func main() {
	ctx := context.Background()

	fmt.Println("=== Layer-Based Error Handling ===")

	// =============================
	// 1. REPOSITORY LAYER
	// =============================
	fmt.Println("1. Repository Layer Errors:")
	repoErr := simulateRepositoryError(ctx, "user123")
	fmt.Printf("   Type: %s\n", repoErr.Type)
	fmt.Printf("   Code: %s\n", repoErr.Code)
	fmt.Printf("   Message: %s\n", repoErr.UserMsg)
	fmt.Printf("   HTTP: %d\n\n", repoErr.HTTPStatus)

	// =============================
	// 2. SERVICE LAYER
	// =============================
	fmt.Println("2. Service Layer - Mapping Repo Error:")
	serviceErr := simulateServiceError(ctx, repoErr)
	fmt.Printf("   Type: %s\n", serviceErr.Type)
	fmt.Printf("   Code: %s\n", serviceErr.Code)
	fmt.Printf("   Message: %s\n", serviceErr.UserMsg)
	fmt.Printf("   HTTP: %d\n\n", serviceErr.HTTPStatus)

	// =============================
	// 3. HANDLER LAYER
	// =============================
	fmt.Println("3. Handler Layer - Mapping Service Error:")
	handlerErr := simulateHandlerError(ctx, serviceErr)
	fmt.Printf("   Type: %s\n", handlerErr.Type)
	fmt.Printf("   Code: %s\n", handlerErr.Code)
	fmt.Printf("   Message: %s\n", handlerErr.UserMsg)
	fmt.Printf("   HTTP: %d\n\n", handlerErr.HTTPStatus)

	fmt.Println("=== All Layer Error Codes ===")

	fmt.Println("Repository Layer (2xxx):")
	fmt.Println("  REPO_NOT_FOUND          -> 2001 -> HTTP 404")
	fmt.Println("  REPO_ALREADY_EXISTS     -> 2002 -> HTTP 409")
	fmt.Println("  REPO_DATABASE_ERROR     -> 2003 -> HTTP 500")
	fmt.Println("  REPO_TIMEOUT            -> 2004 -> HTTP 504")

	fmt.Println("\nService Layer (3xxx):")
	fmt.Println("  SERVICE_INVALID_INPUT   -> 3001 -> HTTP 400")
	fmt.Println("  SERVICE_VALIDATION      -> 3002 -> HTTP 400")
	fmt.Println("  SERVICE_NOT_FOUND       -> 3003 -> HTTP 404")
	fmt.Println("  SERVICE_ALREADY_EXISTS  -> 3004 -> HTTP 409")
	fmt.Println("  SERVICE_UNAUTHORIZED    -> 3005 -> HTTP 401")
	fmt.Println("  SERVICE_FORBIDDEN       -> 3006 -> HTTP 403")

	fmt.Println("\nHandler Layer (4xxx, 5xxx):")
	fmt.Println("  HANDLER_BAD_REQUEST     -> 4000 -> HTTP 400")
	fmt.Println("  HANDLER_UNAUTHORIZED    -> 4001 -> HTTP 401")
	fmt.Println("  HANDLER_FORBIDDEN       -> 4003 -> HTTP 403")
	fmt.Println("  HANDLER_NOT_FOUND       -> 4004 -> HTTP 404")
	fmt.Println("  HANDLER_CONFLICT        -> 4009 -> HTTP 409")
	fmt.Println("  HANDLER_INTERNAL_ERROR  -> 5000 -> HTTP 500")
}

// simulateRepositoryError simulates a repository layer error
func simulateRepositoryError(ctx context.Context, userID string) *errors.AppError {
	// Repository layer creates REPO_* errors
	return errors.NewRepoError(ctx, errors.ErrRepoNotFound, "user record not found in database").
		WithField("user_id", userID).
		WithField("table", "users").
		WithDetails("The user with the specified ID does not exist in the database")
}

// simulateServiceError simulates service layer handling repository error
func simulateServiceError(ctx context.Context, repoErr error) *errors.AppError {
	// Service layer maps repository errors to SERVICE_* errors
	return errors.MapRepoToServiceError(ctx, repoErr)
}

// simulateHandlerError simulates handler layer handling service error
func simulateHandlerError(ctx context.Context, serviceErr error) *errors.AppError {
	// Handler layer maps service errors to HANDLER_* errors
	return errors.MapServiceToHandlerError(ctx, serviceErr)
}
