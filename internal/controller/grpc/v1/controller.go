package v1

import (
	"github.com/go-playground/validator/v10"

	"gct/internal/usecase"
	"gct/pkg/logger"
)

type V1 struct {
	u *usecase.UseCase
	l logger.Log
	v *validator.Validate
}
