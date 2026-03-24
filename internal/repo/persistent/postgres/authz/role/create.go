package role

import (
	"context"
	"fmt"
	"time"

	"gct/internal/shared/domain/consts"
	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (r *Repo) Create(ctx context.Context, role *domain.Role) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns("name", "description", "created_at").
		Values(role.Name, role.Description, time.Now()).
		Suffix(fmt.Sprintf("RETURNING %s, %s", "id", "created_at")).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&role.ID, &role.CreatedAt)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
