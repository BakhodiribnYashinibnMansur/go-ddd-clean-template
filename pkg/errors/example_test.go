package errors_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/evrone/go-clean-template/pkg/errors"
)

func TestBasicUsage(t *testing.T) {
	ctx := context.Background()

	// Oddiy error yaratish
	err1 := errors.New(ctx, errors.ErrUserNotFound, "user not found")
	fmt.Printf("Error 1: %v\n", err1)

	// Field bilan
	err2 := errors.New(ctx, errors.ErrUserNotFound, "user not found").
		WithField("user_id", "123").
		WithField("email", "test@example.com")
	fmt.Printf("Error 2: %v\n", err2)
	fmt.Printf("Fields: %+v\n", err2.Fields)

	// Wrapping
	baseErr := fmt.Errorf("database connection failed")
	err3 := errors.Wrap(ctx, baseErr, errors.ErrDatabase, "failed to connect to db")
	fmt.Printf("Error 3: %v\n", err3)

	// Checking
	if errors.Is(err1, errors.ErrUserNotFound) {
		fmt.Println("✓ Error is USER_NOT_FOUND")
	}

	// HTTP status
	fmt.Printf("HTTP Status: %d\n", err1.HTTPStatus)
	fmt.Printf("User Message: %s\n", err1.UserMsg)
}

func ExampleNew() {
	ctx := context.Background()

	err := errors.New(ctx, errors.ErrUserNotFound, "user not found")
	fmt.Println(err.Type)
	fmt.Println(err.HTTPStatus)
	// Output:
	// USER_NOT_FOUND
	// 404
}

func ExampleWrap() {
	ctx := context.Background()

	baseErr := fmt.Errorf("connection timeout")
	err := errors.Wrap(ctx, baseErr, errors.ErrDatabase, "failed to query user")

	fmt.Println(err.Type)
	// Output:
	// DATABASE_ERROR
}

func ExampleAppError_WithField() {
	ctx := context.Background()

	err := errors.New(ctx, errors.ErrValidation, "validation failed").
		WithField("field", "email").
		WithField("reason", "invalid format")

	fmt.Println(err.Fields["field"])
	fmt.Println(err.Fields["reason"])
	// Output:
	// email
	// invalid format
}
