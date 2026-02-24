package usersetting

import (
	"context"
	"time"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Set(ctx context.Context, userID uuid.UUID, key, value string) error {
	s := &domain.UserSetting{
		ID:        uuid.New(),
		UserID:    userID,
		Key:       key,
		Value:     value,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.repo.Upsert(ctx, s); err != nil {
		uc.logger.Errorw("usersetting.Set failed", "user_id", userID, "key", key, "error", err)
		return err
	}
	return nil
}
