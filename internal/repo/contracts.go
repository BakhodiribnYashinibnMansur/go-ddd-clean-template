// Package repo implements application outer layer logic. Each logic group in own file.
package repo

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/google/uuid"
)

//go:generate mockgen -source=contracts.go -destination=../usecase/mocks_repo_test.go -package=usecase_test

type (
	// TranslationRepo -.
	TranslationRepo interface {
		Store(context.Context, entity.Translation) error
		GetHistory(context.Context) ([]entity.Translation, error)
	}

	// TranslationWebAPI -.
	TranslationWebAPI interface {
		Translate(entity.Translation) (entity.Translation, error)
	}

	// UserRepo -.
	UserRepo interface {
		Create(ctx context.Context, u entity.User) error
		GetByID(ctx context.Context, id int64) (entity.User, error)
		GetByPhone(ctx context.Context, phone string) (entity.User, error)
		Update(ctx context.Context, u entity.User) error
		Delete(ctx context.Context, id int64) error
	}

	// SessionRepo -.
	SessionRepo interface {
		Create(ctx context.Context, s entity.Session) (entity.Session, error)
		GetByID(ctx context.Context, id uuid.UUID) (entity.Session, error)
		GetByUserID(ctx context.Context, turonID int64) ([]entity.Session, error)
		GetByDeviceID(ctx context.Context, turonID int64, deviceID uuid.UUID) (entity.Session, error)
		UpdateActivity(ctx context.Context, id uuid.UUID, fcmToken *string) error
		Delete(ctx context.Context, id uuid.UUID) error
		DeleteByUserID(ctx context.Context, turonID int64) error
		DeleteExpired(ctx context.Context) error
	}
)
