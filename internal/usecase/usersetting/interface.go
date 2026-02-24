package usersetting

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

type UseCaseI interface {
	Gets(ctx context.Context, userID uuid.UUID) ([]domain.UserSetting, error)
	Set(ctx context.Context, userID uuid.UUID, key, value string) error
	Delete(ctx context.Context, userID uuid.UUID, key string) error
	SetPasscode(ctx context.Context, userID uuid.UUID, passcode string) error
	VerifyPasscode(ctx context.Context, userID uuid.UUID, passcode string) (bool, error)
	RemovePasscode(ctx context.Context, userID uuid.UUID) error
}
