package v1

import (
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/go-playground/validator/v10"
)

type V1 struct {
	u *usecase.UseCase
	l logger.Log
	v *validator.Validate
}
