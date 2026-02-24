package usersetting

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) Gets(ctx context.Context, userID uuid.UUID) ([]domain.UserSetting, error) {
	settings, err := uc.repo.Gets(ctx, userID)
	if err != nil {
		uc.logger.Errorw("usersetting.Gets failed", "user_id", userID, "error", err)
		return nil, err
	}
	return settings, nil
}
