package relation

import (
	"context"
	"fmt"
	"time"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, relation *domain.Relation) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns(schema.RelationType, schema.RelationName, schema.RelationCreatedAt).
		Values(relation.Type, relation.Name, time.Now()).
		Suffix(fmt.Sprintf("RETURNING %s, %s", schema.RelationID, schema.RelationCreatedAt)).
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
