package relation

import (
	"context"
	"fmt"
	"time"

	"gct/consts"
	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, relation *domain.Relation) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns("type", "name", "created_at").
		Values(relation.Type, relation.Name, time.Now()).
		Suffix(fmt.Sprintf("RETURNING %s, %s", "id", "created_at")).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&relation.ID, &relation.CreatedAt)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
