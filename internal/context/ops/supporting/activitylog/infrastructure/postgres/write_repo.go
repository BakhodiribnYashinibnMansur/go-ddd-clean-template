package postgres

import (
	"context"

	"gct/internal/context/ops/supporting/activitylog/domain"
	"gct/internal/kernel/consts"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

var activityLogColumns = []string{
	"actor_id", "action", "entity_type", "entity_id",
	"field_name", "old_value", "new_value", "metadata", "request_id", "created_at",
}

// ActivityLogWriteRepo implements domain.ActivityLogWriteRepository using PostgreSQL.
type ActivityLogWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewActivityLogWriteRepo creates a new ActivityLogWriteRepo.
func NewActivityLogWriteRepo(pool *pgxpool.Pool) *ActivityLogWriteRepo {
	return &ActivityLogWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// SaveBatch inserts multiple activity log entries in a single multi-row INSERT.
func (r *ActivityLogWriteRepo) SaveBatch(ctx context.Context, entries []*domain.ActivityLogEntry) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "ActivityLogWriteRepo.SaveBatch")
	defer func() { end(err) }()

	if len(entries) == 0 {
		return nil
	}

	qb := r.builder.
		Insert(consts.TableActivityLog).
		Columns(activityLogColumns...)

	for _, e := range entries {
		qb = qb.Values(
			e.ActorID(),
			e.Action(),
			e.EntityType(),
			e.EntityID(),
			e.FieldName(),
			e.OldValue(),
			e.NewValue(),
			e.Metadata(),
			e.RequestID(),
			e.CreatedAt(),
		)
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = pgxutil.QuerierFromContext(ctx, r.pool).Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, consts.TableActivityLog, nil)
	}

	return nil
}
