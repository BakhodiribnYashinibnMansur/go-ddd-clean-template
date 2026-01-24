package sitesetting

import (
	"context"

	"gct/internal/domain"
	"gct/internal/repo/persistent"
	"gct/pkg/logger"
)

type UseCase struct {
	repo   *persistent.Repo
	logger logger.Log
}

func New(repo *persistent.Repo, logger logger.Log) *UseCase {
	return &UseCase{
		repo:   repo,
		logger: logger,
	}
}

// Get retrieves a single site setting by filter
func (uc *UseCase) Get(ctx context.Context, filter *domain.SiteSettingFilter) (*domain.SiteSetting, error) {
	setting, err := uc.repo.Postgres.SiteSetting.Get(ctx, filter)
	if err != nil {
		uc.logger.WithContext(ctx).Errorw("failed to get site setting", "filter", filter, "error", err)
		return nil, err
	}
	return setting, nil
}

// Gets retrieves multiple site settings with filters
func (uc *UseCase) Gets(ctx context.Context, filter *domain.SiteSettingsFilter) ([]*domain.SiteSetting, int, error) {
	settings, count, err := uc.repo.Postgres.SiteSetting.Gets(ctx, filter)
	if err != nil {
		uc.logger.WithContext(ctx).Errorw("failed to get site settings", "filter", filter, "error", err)
		return nil, 0, err
	}
	return settings, count, nil
}

// Update updates a site setting
func (uc *UseCase) Update(ctx context.Context, setting *domain.SiteSetting) error {
	err := uc.repo.Postgres.SiteSetting.Update(ctx, setting)
	if err != nil {
		uc.logger.WithContext(ctx).Errorw("failed to update site setting", "setting_id", setting.ID, "error", err)
		return err
	}
	uc.logger.WithContext(ctx).Infow("site setting updated", "setting_id", setting.ID, "key", setting.Key)
	return nil
}

// UpdateByKey updates setting value by key
func (uc *UseCase) UpdateByKey(ctx context.Context, key, value string) error {
	err := uc.repo.Postgres.SiteSetting.UpdateByKey(ctx, key, value)
	if err != nil {
		uc.logger.WithContext(ctx).Errorw("failed to update site setting by key", "key", key, "error", err)
		return err
	}
	uc.logger.WithContext(ctx).Infow("site setting updated by key", "key", key)
	return nil
}

// GetByKey is a convenience method to get a setting by its key
func (uc *UseCase) GetByKey(ctx context.Context, key string) (*domain.SiteSetting, error) {
	return uc.Get(ctx, &domain.SiteSettingFilter{Key: &key})
}
