package policy

import (
	"context"
	"fmt"
	"time"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, p *domain.Policy) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns(
			schema.PolicyPermissionID,
			schema.PolicyEffect,
			schema.PolicyPriority,
			schema.PolicyActive,
			schema.PolicyConditions,
			schema.PolicyCreatedAt,
		).
		Values(p.PermissionID, p.Effect, p.Priority, p.Active, p.Conditions, time.Now()).
		Suffix(fmt.Sprintf("RETURNING %s, %s", schema.PolicyID, schema.PolicyCreatedAt)).
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
