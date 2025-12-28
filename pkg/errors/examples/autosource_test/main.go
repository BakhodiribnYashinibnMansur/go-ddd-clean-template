package main

import (
	"context"
	"fmt"

	apperrors "github.com/evrone/go-clean-template/pkg/errors"
)

// Test function to show AutoSource works
func TestAutoSource() {
	ctx := context.Background()

	// Create an error
	err := apperrors.AutoSource(
		apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
			"test error")).
		WithField("table", "users")

	// Print the error fields
	fmt.Printf("Error Type: %s\n", err.Type)
	fmt.Printf("Error Code: %s\n", err.Code)
	fmt.Printf("Error Message: %s\n", err.Message)
	fmt.Println("\nFields:")
	for key, value := range err.Fields {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

func main() {
	TestAutoSource()
}
