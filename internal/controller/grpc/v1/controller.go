package v1

import (
	"gct/internal/usecase"
	"gct/pkg/logger"
	"github.com/go-playground/validator/v10"
)

type V1 struct {
	u *usecase.UseCase
	l logger.Log
	v *validator.Validate
}
