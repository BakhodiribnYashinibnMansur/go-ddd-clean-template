// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/google/uuid"
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

	// Session -.
	Session interface {
		Create(ctx context.Context, s entity.Session, duration time.Duration) (entity.Session, error)
		GetByID(ctx context.Context, id uuid.UUID) (entity.Session, error)
		GetByUserID(ctx context.Context, turonID int64) ([]entity.Session, error)
		GetOrCreateByDevice(ctx context.Context, turonID int64, deviceID uuid.UUID, s entity.Session, duration time.Duration) (entity.Session, error)
		UpdateActivity(ctx context.Context, id uuid.UUID, fcmToken *string) error
		Delete(ctx context.Context, id uuid.UUID) error
		DeleteByUserID(ctx context.Context, turonID int64) error
		CleanupExpired(ctx context.Context) error
	}
)
