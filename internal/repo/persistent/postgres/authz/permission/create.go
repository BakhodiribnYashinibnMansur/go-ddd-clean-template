package permission

import (
	"context"
	"fmt"
	"time"

	"gct/internal/shared/domain/consts"
	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (r *Repo) Create(ctx context.Context, p *domain.Permission) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns("parent_id", "name", "description", "created_at").
		Values(p.ParentID, p.Name, p.Description, time.Now()).
		Suffix(fmt.Sprintf("RETURNING %s, %s", "id", "created_at")).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
