package postgres

import (
	"context"
	"time"

	"gct/internal/iprule/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableIPRules

var writeColumns = []string{
	"id", "ip_address", "action", "reason", "expires_at", "created_at", "updated_at",
}

// IPRuleWriteRepo implements domain.IPRuleRepository using PostgreSQL.
type IPRuleWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewIPRuleWriteRepo creates a new IPRuleWriteRepo.
func NewIPRuleWriteRepo(pool *pgxpool.Pool) *IPRuleWriteRepo {
	return &IPRuleWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new IPRule aggregate into the database.
func (r *IPRuleWriteRepo) Save(ctx context.Context, rule *domain.IPRule) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			rule.ID(), rule.IPAddress(), rule.Action(), rule.Reason(),
			rule.ExpiresAt(), rule.CreatedAt(), rule.UpdatedAt(),
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// FindByID retrieves an IPRule aggregate by its ID.
func (r *IPRuleWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.IPRule, error) {
	sql, args, err := r.builder.
		Select(writeColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanIPRule(row)
}

// Update updates an existing IPRule aggregate in the database.
func (r *IPRuleWriteRepo) Update(ctx context.Context, rule *domain.IPRule) error {
	sql, args, err := r.builder.
		Update(tableName).
		Set("ip_address", rule.IPAddress()).
		Set("action", rule.Action()).
		Set("reason", rule.Reason()).
		Set("expires_at", rule.ExpiresAt()).
		Set("updated_at", rule.UpdatedAt()).
		Where(squirrel.Eq{"id": rule.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// Delete removes an IPRule by its ID.
func (r *IPRuleWriteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.builder.
		Delete(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// List retrieves a paginated list of IPRule aggregates with optional filters.
func (r *IPRuleWriteRepo) List(ctx context.Context, filter domain.IPRuleFilter) ([]*domain.IPRule, int64, error) {
	conds := squirrel.And{}
	conds = applyFilters(conds, filter)

	countQB := r.builder.Select("COUNT(*)").From(tableName)
	if len(conds) > 0 {
		countQB = countQB.Where(conds)
	}
	countSQL, countArgs, err := countQB.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var total int64
	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	qb := r.builder.
		Select(writeColumns...).
		From(tableName).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(filter.Offset))

	if len(conds) > 0 {
		qb = qb.Where(conds)
	}

	sqlStr, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}
	defer rows.Close()

	var results []*domain.IPRule
	for rows.Next() {
		rule, err := scanIPRuleFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		results = append(results, rule)
	}

	return results, total, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func applyFilters(conds squirrel.And, filter domain.IPRuleFilter) squirrel.And {
	if filter.IPAddress != nil {
		conds = append(conds, squirrel.Eq{"ip_address": *filter.IPAddress})
	}
	if filter.Action != nil {
		conds = append(conds, squirrel.Eq{"action": *filter.Action})
	}
	return conds
}

func scanIPRule(row pgx.Row) (*domain.IPRule, error) {
	var (
		id        uuid.UUID
		ipAddress string
		action    string
		reason    string
		expiresAt *time.Time
		createdAt time.Time
		updatedAt time.Time
	)

	err := row.Scan(&id, &ipAddress, &action, &reason, &expiresAt, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return domain.ReconstructIPRule(id, createdAt, updatedAt, ipAddress, action, reason, expiresAt), nil
}

func scanIPRuleFromRows(rows pgx.Rows) (*domain.IPRule, error) {
	var (
		id        uuid.UUID
		ipAddress string
		action    string
		reason    string
		expiresAt *time.Time
		createdAt time.Time
		updatedAt time.Time
	)

	err := rows.Scan(&id, &ipAddress, &action, &reason, &expiresAt, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	return domain.ReconstructIPRule(id, createdAt, updatedAt, ipAddress, action, reason, expiresAt), nil
}
