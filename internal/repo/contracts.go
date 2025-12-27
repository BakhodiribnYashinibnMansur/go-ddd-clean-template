// Package repo implements application outer layer logic. Each logic group in own file.
package repo

import (
	"context"

	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/google/uuid"
)

//go:generate mockgen -source=contracts.go -destination=../usecase/mocks_repo_test.go -package=usecase_test

type (
	// TranslationRepo -.
	TranslationRepo interface {
		Store(context.Context, domain.Translation) error
		GetHistory(context.Context) ([]domain.Translation, error)
	}

	// TranslationWebAPI -.
	TranslationWebAPI interface {
		Translate(domain.Translation) (domain.Translation, error)
	}

	// UserRepo -.
	UserRepo interface {
		Create(ctx context.Context, u domain.User) error
		GetByID(ctx context.Context, id int64) (domain.User, error)
		GetByPhone(ctx context.Context, phone string) (domain.User, error)
		Update(ctx context.Context, u domain.User) error
		Delete(ctx context.Context, id int64) error
	}

	// SessionRepo -.
	SessionRepo interface {
		Create(ctx context.Context, s domain.Session) (domain.Session, error)
		GetByID(ctx context.Context, id uuid.UUID) (domain.Session, error)
		GetByUserID(ctx context.Context, turonID int64) ([]domain.Session, error)
		GetByDeviceID(ctx context.Context, turonID int64, deviceID uuid.UUID) (domain.Session, error)
		UpdateActivity(ctx context.Context, id uuid.UUID, fcmToken *string) error
		Delete(ctx context.Context, id uuid.UUID) error
		DeleteByUserID(ctx context.Context, turonID int64) error
		DeleteExpired(ctx context.Context) error
	}
)
