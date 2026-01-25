package permission

import (
	"context"
	"fmt"
	"time"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, p *domain.Permission) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns(schema.PermissionParentID, schema.PermissionName, schema.PermissionCreatedAt).
		Values(p.ParentID, p.Name, time.Now()).
		Suffix(fmt.Sprintf("RETURNING %s, %s", schema.PermissionID, schema.PermissionCreatedAt)).
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
