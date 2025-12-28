// Package main - Error handling tizimini amaliy misol
package main

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/pkg/errors"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== 1. Oddiy Error Yaratish ===")
	err1 := errors.New(ctx, errors.ErrUserNotFound, "user not found")
	fmt.Printf("%v\n\n", err1)

	fmt.Println("=== 2. Field bilan ===")
	err2 := errors.New(ctx, errors.ErrUserNotFound, "user not found").
		WithField("user_id", "123").
		WithField("phone", "+998901234567")
	fmt.Printf("%v\n", err2)
	fmt.Printf("HTTP status: %d\n", err2.HTTPStatus)
	fmt.Printf("User message: %s\n", err2.UserMsg)
	fmt.Printf("Fields: %+v\n\n", err2.Fields)

	fmt.Println("=== 3. Error Wrapping ===")
	// Masalan, database error
	dbErr := fmt.Errorf("connection timeout")
	wrappedErr := errors.Wrap(ctx, dbErr, errors.ErrDatabase, "failed to connect to database")
	fmt.Printf("%v\n\n", wrappedErr)

	fmt.Println("=== 4. Error Checking ===")
	if errors.Is(err1, errors.ErrUserNotFound) {
		fmt.Println("✓ Error is USER_NOT_FOUND")
	}

	code := errors.GetCode(err2)
	fmt.Printf("Error code: %s\n\n", code)

	fmt.Println("=== 5. Validation Example ===")
	validationErr := errors.New(ctx, errors.ErrValidation, "validation failed").
		WithField("field", "email").
		WithField("reason", "invalid format").
		WithField("value", "not-an-email")
	fmt.Printf("%v\n", validationErr)
	fmt.Printf("Validation details: %+v\n\n", validationErr.Fields)

	fmt.Println("=== 6. Auth Error Example ===")
	authErr := errors.New(ctx, errors.ErrExpiredToken, "token expired").
		WithField("token_id", "abc123").
		WithField("expired_at", "2024-12-28")
	fmt.Printf("%v\n", authErr)
	fmt.Printf("HTTP: %d, Message: %s\n\n", authErr.HTTPStatus, authErr.UserMsg)

	fmt.Println("=== 7. Real Use Case - Repository Layer ===")
	user, err := simulateRepositoryCall(ctx, "user123")
	if err != nil {
		fmt.Printf("Repository error: %v\n", err)
		if appErr, ok := err.(*errors.AppError); ok {
			fmt.Printf("  Code: %s\n", appErr.Code)
			fmt.Printf("  HTTP: %d\n", appErr.HTTPStatus)
			fmt.Printf("  Fields: %+v\n", appErr.Fields)
		}
	} else {
		fmt.Printf("User: %+v\n", user)
	}
}

// simulateRepositoryCall repository layerni simulyatsiya qiladi
func simulateRepositoryCall(ctx context.Context, userID string) (map[string]any, error) {
	// Bu yerda GORM dan error kelgan deb o'ylaymiz
	dbErr := fmt.Errorf("record not found")

	// Uni wrap qilamiz
	return nil, errors.Wrap(ctx, dbErr, errors.ErrUserNotFound, "user not found in database").
		WithField("user_id", userID).
		WithField("table", "users")
}
