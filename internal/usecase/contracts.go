// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"
	"time"

	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/google/uuid"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_usecase_test.go -package=usecase_test

type (
	// Translation -.
	Translation interface {
		Translate(context.Context, domain.Translation) (domain.Translation, error)
		History(context.Context) (domain.TranslationHistory, error)
	}

	// User -.
	User interface {
		Create(context.Context, domain.User) error
		GetByID(context.Context, int64) (domain.User, error)
		GetByPhone(context.Context, string) (domain.User, error)
		Update(context.Context, domain.User) error
		Delete(context.Context, int64) error
	}

	// Session -.
	Session interface {
		Create(ctx context.Context, s domain.Session, duration time.Duration) (domain.Session, error)
		GetByID(ctx context.Context, id uuid.UUID) (domain.Session, error)
		GetByUserID(ctx context.Context, turonID int64) ([]domain.Session, error)
		GetOrCreateByDevice(ctx context.Context, turonID int64, deviceID uuid.UUID, s domain.Session, duration time.Duration) (domain.Session, error)
		UpdateActivity(ctx context.Context, id uuid.UUID, fcmToken *string) error
		Delete(ctx context.Context, id uuid.UUID) error
		DeleteByUserID(ctx context.Context, turonID int64) error
		CleanupExpired(ctx context.Context) error
	}
)
