package validator

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	apperrors "gct/pkg/errors"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register function to get tag name from json tag
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// ValidateStruct validates a struct and returns a standardized AppError if validation fails
func ValidateStruct(ctx context.Context, s any) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	if errs, ok := err.(validator.ValidationErrors); ok {
		var details []string
		for _, e := range errs {
			details = append(details, fmt.Sprintf("[%s]: %s %s", e.Field(), e.Tag(), e.Param()))
		}

		msg := "Validation failed: " + strings.Join(details, ", ")
		return apperrors.NewValidationError(ctx, msg).
			WithDetails(strings.Join(details, "\n")).
			WithField("validation_errors", details)
	}

	return apperrors.NewInternalError(ctx, "Validation process failed: "+err.Error())
}
