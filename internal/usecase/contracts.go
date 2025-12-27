// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_usecase_test.go -package=usecase_test

type (
	// Translation -.
	Translation interface {
		Translate(context.Context, entity.Translation) (entity.Translation, error)
		History(context.Context) (entity.TranslationHistory, error)
	}

	// User -.
	User interface {
		Create(context.Context, entity.User) error
		GetByID(context.Context, int64) (entity.User, error)
		GetByPhone(context.Context, string) (entity.User, error)
		Update(context.Context, entity.User) error
		Delete(context.Context, int64) error
	}
)
