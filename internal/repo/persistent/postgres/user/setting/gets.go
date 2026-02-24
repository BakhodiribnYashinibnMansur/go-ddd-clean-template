package setting

import (
	"context"

	"gct/consts"
	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) Gets(ctx context.Context, userID uuid.UUID) ([]domain.UserSetting, error) {
	sql, args, err := r.builder.
		Select("id", "user_id", "key", "value", "created_at", "updated_at").
		From(tableName).
		Where(squirrel.Eq{"user_id": userID}).
		OrderBy("key " + consts.SQLOrderAsc).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}
	defer rows.Close()

	var settings []domain.UserSetting
	for rows.Next() {
		var s domain.UserSetting
		if err := rows.Scan(&s.ID, &s.UserID, &s.Key, &s.Value, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToScanRow)
		}
		settings = append(settings, s)
	}
	return settings, rows.Err()
}
