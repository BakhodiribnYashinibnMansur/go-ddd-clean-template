package setting

import (
	"context"
	"time"

	"gct/internal/shared/domain/consts"
	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (r *Repo) Upsert(ctx context.Context, s *domain.UserSetting) error {
	now := time.Now()
	sql, args, err := r.builder.
		Insert(tableName).
		Columns("id", "user_id", "key", "value", "created_at", "updated_at").
		Values(s.ID, s.UserID, s.Key, s.Value, now, now).
		Suffix("ON CONFLICT (user_id, key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	return apperrors.HandlePgError(err, tableName, nil)
}
