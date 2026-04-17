package postgres

import (
	"context"
	"time"

	"gct/internal/context/ops/supporting/activitylog/domain"
	"gct/internal/kernel/consts"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var readActivityLogColumns = []string{
	"id", "actor_id", "action", "entity_type", "entity_id",
	"field_name", "old_value", "new_value", "metadata", "request_id", "created_at",
}

// ActivityLogReadRepo implements domain.ActivityLogReadRepository using PostgreSQL.
type ActivityLogReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewActivityLogReadRepo creates a new ActivityLogReadRepo.
func NewActivityLogReadRepo(pool *pgxpool.Pool) *ActivityLogReadRepo {
	return &ActivityLogReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// List returns a paginated list of activity log views with optional filters.
func (r *ActivityLogReadRepo) List(ctx context.Context, filter domain.ActivityLogFilter) (items []*domain.ActivityLogView, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "ActivityLogReadRepo.List")
	defer func() { end(err) }()

	conds := squirrel.And{}

	if filter.ActorID != nil {
		conds = append(conds, squirrel.Eq{"actor_id": *filter.ActorID})
	}
	if filter.EntityType != nil {
		conds = append(conds, squirrel.Eq{"entity_type": *filter.EntityType})
	}
	if filter.EntityID != nil {
		conds = append(conds, squirrel.Eq{"entity_id": *filter.EntityID})
	}
	if filter.FieldName != nil {
		conds = append(conds, squirrel.Eq{"field_name": *filter.FieldName})
	}
	if filter.Action != nil {
		conds = append(conds, squirrel.Eq{"action": *filter.Action})
	}
	if filter.FromDate != nil {
		conds = append(conds, squirrel.GtOrEq{"created_at": *filter.FromDate})
	}
	if filter.ToDate != nil {
		conds = append(conds, squirrel.LtOrEq{"created_at": *filter.ToDate})
	}

	// Count total.
	countQB := r.builder.Select("COUNT(*)").From(consts.TableActivityLog)
	if len(conds) > 0 {
		countQB = countQB.Where(conds)
	}

	countSQL, countArgs, err := countQB.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TableActivityLog, nil)
	}

	// Fetch page.
	qb := r.builder.Select(readActivityLogColumns...).From(consts.TableActivityLog)
	if len(conds) > 0 {
		qb = qb.Where(conds)
	}

	if filter.Pagination != nil {
		qb = qb.Limit(uint64(filter.Pagination.Limit)).
			Offset(uint64(filter.Pagination.Offset))

		if ob := filter.Pagination.SafeOrderBy(); ob != "" {
			qb = qb.OrderBy(ob)
		}
	}

	if filter.Pagination == nil || filter.Pagination.SortBy == "" {
		qb = qb.OrderBy("created_at DESC")
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TableActivityLog, nil)
	}
	defer rows.Close()

	var views []*domain.ActivityLogView
	for rows.Next() {
		var (
			id         int64
			actorID    uuid.UUID
			action     string
			entityType string
			entityID   uuid.UUID
			fieldName  *string
			oldValue   *string
			newValue   *string
			metadata   *string
			requestID  *string
			createdAt  time.Time
		)

		if err := rows.Scan(
			&id, &actorID, &action, &entityType, &entityID,
			&fieldName, &oldValue, &newValue, &metadata, &requestID, &createdAt,
		); err != nil {
			return nil, 0, apperrors.HandlePgError(err, consts.TableActivityLog, nil)
		}

		views = append(views, &domain.ActivityLogView{
			ID:         id,
			ActorID:    actorID,
			Action:     action,
			EntityType: entityType,
			EntityID:   entityID,
			FieldName:  fieldName,
			OldValue:   oldValue,
			NewValue:   newValue,
			Metadata:   metadata,
			RequestID:  requestID,
			CreatedAt:  createdAt,
		})
	}

	return views, total, nil
}
