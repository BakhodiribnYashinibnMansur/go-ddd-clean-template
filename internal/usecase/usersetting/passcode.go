package usersetting

import (
	"context"
	"time"

	"gct/internal/domain"

	"github.com/google/uuid"
)

func (uc *UseCase) SetPasscode(ctx context.Context, userID uuid.UUID, passcode string) error {
	now := time.Now()
	if err := uc.repo.Upsert(ctx, &domain.UserSetting{
		ID: uuid.New(), UserID: userID, Key: KeyPasscode,
		Value: passcode, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		return err
	}
	return uc.repo.Upsert(ctx, &domain.UserSetting{
		ID: uuid.New(), UserID: userID, Key: KeyPasscodeEnabled,
		Value: "true", CreatedAt: now, UpdatedAt: now,
	})
}

func (uc *UseCase) VerifyPasscode(ctx context.Context, userID uuid.UUID, passcode string) (bool, error) {
	settings, err := uc.repo.Gets(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, s := range settings {
		if s.Key == KeyPasscode {
			return s.Value == passcode, nil
		}
	}
	return false, nil
}

func (uc *UseCase) RemovePasscode(ctx context.Context, userID uuid.UUID) error {
	_ = uc.repo.Delete(ctx, userID, KeyPasscode)
	_ = uc.repo.Delete(ctx, userID, KeyPasscodeEnabled)
	return nil
}
