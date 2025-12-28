package main

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Initialize Zap logger
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, _ := config.Build()
	defer logger.Sync()

	ctx := context.Background()

	fmt.Println("=== Error Logging Examples ===")

	// =============================
	// 1. Repository Error Logging
	// =============================
	fmt.Println("1. Repository Error:")
	repoErr := errors.NewRepoError(ctx, errors.ErrRepoNotFound, "user not found in database").
		WithField("user_id", "12345").
		WithField("table", "users").
		WithDetails("The user with the ID '12345' does not exist in our database")

	logError(logger, repoErr)

	// =============================
	// 2. Service Error Logging
	// =============================
	fmt.Println("\n2. Service Error:")
	serviceErr := errors.NewServiceError(ctx, errors.ErrServiceValidation, "validation failed").
		WithField("field", "email").
		WithField("value", "invalid-email").
		WithDetails("Email format is invalid")

	logError(logger, serviceErr)

	// =============================
	// 3. Handler Error Logging
	// =============================
	fmt.Println("\n3. Handler Error:")
	handlerErr := errors.NewHandlerError(ctx, errors.ErrHandlerUnauthorized, "authentication required").
		WithDetails("Missing or invalid authentication token")

	logError(logger, handlerErr)

	// =============================
	// 4. Wrapped Error Logging
	// =============================
	fmt.Println("\n4. Wrapped Error:")
	baseErr := fmt.Errorf("connection timeout after 30s")
	wrappedErr := errors.WrapRepoError(ctx, baseErr, errors.ErrRepoTimeout, "database query timeout").
		WithField("query", "SELECT * FROM users WHERE id = ?").
		WithField("timeout", "30s")

	logError(logger, wrappedErr)

	// =============================
	// 5. Complete Error Flow
	// =============================
	fmt.Println("\n5. Complete Error Flow (Repo -> Service -> Handler):")

	// Repository error
	dbErr := errors.NewRepoError(ctx, errors.ErrRepoNotFound, "record not found")
	logger.Error("Repository layer error",
		zap.String("layer", "repository"),
		zap.String("error_type", dbErr.Type),
		zap.String("error_code", dbErr.Code),
		zap.Int("http_status", dbErr.HTTPStatus),
		zap.String("message", dbErr.Message),
	)

	// Service maps it
	svcErr := errors.MapRepoToServiceError(ctx, dbErr)
	logger.Error("Service layer error",
		zap.String("layer", "service"),
		zap.String("error_type", svcErr.Type),
		zap.String("error_code", svcErr.Code),
		zap.Int("http_status", svcErr.HTTPStatus),
		zap.String("message", svcErr.Message),
		zap.String("details", svcErr.Details),
	)

	// Handler maps it
	finalErr := errors.MapServiceToHandlerError(ctx, svcErr)
	logger.Error("Handler layer error",
		zap.String("layer", "handler"),
		zap.String("error_type", finalErr.Type),
		zap.String("error_code", finalErr.Code),
		zap.Int("http_status", finalErr.HTTPStatus),
		zap.String("user_message", finalErr.UserMsg),
		zap.String("details", finalErr.Details),
	)

	fmt.Println("\n=== Log Format Examples ===")
	fmt.Println("Production JSON format:")
	fmt.Println(`{"level":"error","timestamp":"2023-12-08T12:30:45Z","error_type":"REPO_NOT_FOUND","error_code":"2001","http_status":404,"message":"user not found in database","user_id":"12345","table":"users"}`)

	fmt.Println("\nDevelopment console format:")
	fmt.Println(`ERROR  user not found in database  {"error_type": "REPO_NOT_FOUND", "error_code": "2001", "http_status": 404, "user_id": "12345", "table": "users"}`)
}

// logError logs error with all available information
func logError(logger *zap.Logger, err *errors.AppError) {
	fields := []zap.Field{
		zap.String("error_type", err.Type),
		zap.String("error_code", err.Code),
		zap.Int("http_status", err.HTTPStatus),
		zap.String("user_message", err.UserMsg),
	}

	// Add details if present
	if err.Details != "" {
		fields = append(fields, zap.String("details", err.Details))
	}

	// Add custom fields
	for key, value := range err.Fields {
		fields = append(fields, zap.Any(key, value))
	}

	// Add wrapped error if present
	if err.Err != nil {
		fields = append(fields, zap.Error(err.Err))
	}

	logger.Error(err.Message, fields...)
}
