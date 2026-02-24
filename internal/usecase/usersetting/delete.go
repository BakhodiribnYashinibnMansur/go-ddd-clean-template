package usersetting

import (
	"context"

	"github.com/google/uuid"
)

func (uc *UseCase) Delete(ctx context.Context, userID uuid.UUID, key string) error {
	if err := uc.repo.Delete(ctx, userID, key); err != nil {
		uc.logger.Errorw("usersetting.Delete failed", "user_id", userID, "key", key, "error", err)
		return err
	}
	return nil
}
