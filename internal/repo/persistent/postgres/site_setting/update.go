package sitesetting

import (
	"context"
	"time"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Update(ctx context.Context, setting *domain.SiteSetting) error {
	setting.UpdatedAt = time.Now()

	sql := `
		UPDATE site_settings 
		SET value = $1, value_type = $2, category = $3, description = $4, is_public = $5, updated_at = $6
		WHERE id = $7
	`

	_, err := r.pool.Exec(ctx, sql,
		setting.Value,
		setting.ValueType,
		setting.Category,
		setting.Description,
		setting.IsPublic,
		setting.UpdatedAt,
		setting.ID,
	)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// UpdateByKey updates a setting by its key (useful for simple value updates)
func (r *Repo) UpdateByKey(ctx context.Context, key, value string) error {
	sql := `
		UPDATE site_settings 
		SET value = $1, updated_at = $2
		WHERE key = $3
	`

	_, err := r.pool.Exec(ctx, sql, value, time.Now(), key)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
