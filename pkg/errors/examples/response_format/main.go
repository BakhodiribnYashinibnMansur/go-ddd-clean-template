package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/evrone/go-clean-template/pkg/errors"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== Response Format Examples ===")

	// =============================
	// 1. Repository Layer Error
	// =============================
	fmt.Println("\n1. Repository Layer Error Response:")

	repoErr := errors.NewRepoError(ctx, errors.ErrRepoNotFound, "user not found in database").
		WithField("user_id", "12345").
		WithField("table", "users").
		WithDetails("The user with the ID '12345' does not exist in our database records.")

	repoResponse := buildErrorResponse(repoErr, "/api/v1/users/12345", "GET")
	printJSON(repoResponse)

	// =============================
	// 2. Service Layer Error
	// =============================
	fmt.Println("\n2. Service Layer Error Response:")

	serviceErr := errors.NewServiceError(ctx, errors.ErrServiceValidation, "validation failed").
		WithField("field", "email").
		WithField("value", "invalid-email").
		WithDetails("The email address format is invalid. Please provide a valid email address.")

	serviceResponse := buildErrorResponse(serviceErr, "/api/v1/users", "POST")
	printJSON(serviceResponse)

	// =============================
	// 3. Handler Layer Error
	// =============================
	fmt.Println("\n3. Handler Layer Error Response:")

	handlerErr := errors.NewHandlerError(ctx, errors.ErrHandlerUnauthorized, "authentication required").
		WithDetails("You must be authenticated to access this resource. Please provide a valid token.")

	handlerResponse := buildErrorResponse(handlerErr, "/api/v1/profile", "GET")
	printJSON(handlerResponse)

	// =============================
	// 4. Mapped Error Flow
	// =============================
	fmt.Println("\n4. Complete Error Flow (Repo -> Service -> Handler):")

	// Start with repo error
	dbErr := errors.NewRepoError(ctx, errors.ErrRepoNotFound, "record not found")

	// Service maps it
	svcErr := errors.MapRepoToServiceError(ctx, dbErr)

	// Handler maps it
	finalErr := errors.MapServiceToHandlerError(ctx, svcErr)

	flowResponse := buildErrorResponse(finalErr, "/api/v1/resources/999", "DELETE")
	printJSON(flowResponse)

	// =============================
	// 5. Success Response
	// =============================
	fmt.Println("\n5. Success Response:")

	successResponse := map[string]any{
		"status":     "success",
		"statusCode": 200,
		"data": map[string]any{
			"id":    "12345",
			"name":  "John Doe",
			"email": "john@example.com",
		},
	}
	printJSON(successResponse)

	// =============================
	// 6. Error Codes Summary
	// =============================
	fmt.Println("\n=== Error Codes by Layer ===")

	fmt.Println("\nRepository (2xxx):")
	printErrorCode("REPO_NOT_FOUND", "2001", 404)
	printErrorCode("REPO_ALREADY_EXISTS", "2002", 409)
	printErrorCode("REPO_DATABASE_ERROR", "2003", 500)
	printErrorCode("REPO_TIMEOUT", "2004", 504)

	fmt.Println("\nService (3xxx):")
	printErrorCode("SERVICE_INVALID_INPUT", "3001", 400)
	printErrorCode("SERVICE_VALIDATION", "3002", 400)
	printErrorCode("SERVICE_NOT_FOUND", "3003", 404)
	printErrorCode("SERVICE_UNAUTHORIZED", "3005", 401)

	fmt.Println("\nHandler (4xxx, 5xxx):")
	printErrorCode("HANDLER_BAD_REQUEST", "4000", 400)
	printErrorCode("HANDLER_UNAUTHORIZED", "4001", 401)
	printErrorCode("HANDLER_NOT_FOUND", "4004", 404)
	printErrorCode("HANDLER_INTERNAL_ERROR", "5000", 500)
}

// buildErrorResponse builds error response in the required format
func buildErrorResponse(err *errors.AppError, path, method string) map[string]any {
	return map[string]any{
		"status":     "error",
		"statusCode": err.HTTPStatus,
		"error": map[string]any{
			"code":      err.Code,    // Numeric code (e.g., "2001")
			"message":   err.UserMsg, // User-friendly message
			"type":      err.Type,    // Error type (e.g., "REPO_NOT_FOUND")
			"details":   err.Details, // Detailed explanation
			"timestamp": "2023-12-08T12:30:45Z",
			"path":      path,
			"method":    method,
		},
	}
}

// printJSON prints JSON with formatting
func printJSON(data any) {
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(jsonData))
}

// printErrorCode prints error code info
func printErrorCode(errType, code string, httpStatus int) {
	fmt.Printf("  %-30s | Code: %-6s | HTTP: %d\n", errType, code, httpStatus)
}
