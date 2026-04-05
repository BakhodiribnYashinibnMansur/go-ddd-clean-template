package postgres

import (
	"context"
	"time"

	"gct/internal/context/ops/iprule/domain"
	"gct/internal/kernel/consts"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableIPRules

var writeColumns = []string{
	"id", "ip_address", "type", "reason", "is_active", "created_at", "updated_at",
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
func (r *IPRuleWriteRepo) Save(ctx context.Context, rule *domain.IPRule) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "IPRuleWriteRepo.Save")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			rule.ID(), rule.IPAddress(), rule.Action(), rule.Reason(),
			true, rule.CreatedAt(), rule.UpdatedAt(),
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
func (r *IPRuleWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (result *domain.IPRule, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "IPRuleWriteRepo.FindByID")
	defer func() { end(err) }()

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
func (r *IPRuleWriteRepo) Update(ctx context.Context, rule *domain.IPRule) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "IPRuleWriteRepo.Update")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Update(tableName).
		Set("ip_address", rule.IPAddress()).
		Set("type", rule.Action()).
		Set("reason", rule.Reason()).
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
func (r *IPRuleWriteRepo) Delete(ctx context.Context, id uuid.UUID) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "IPRuleWriteRepo.Delete")
	defer func() { end(err) }()

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
func (r *IPRuleWriteRepo) List(ctx context.Context, filter domain.IPRuleFilter) (results []*domain.IPRule, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "IPRuleWriteRepo.List")
	defer func() { end(err) }()

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
		conds = append(conds, squirrel.Eq{"type": *filter.Action})
	}
	return conds
}

func scanIPRule(row pgx.Row) (*domain.IPRule, error) {
	var (
		id        uuid.UUID
		ipAddress string
		ruleType  string
		reason    string
		isActive  bool
		createdAt time.Time
		updatedAt time.Time
	)

	err := row.Scan(&id, &ipAddress, &ruleType, &reason, &isActive, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	_ = isActive

	return domain.ReconstructIPRule(id, createdAt, updatedAt, ipAddress, ruleType, reason, nil), nil
}

func scanIPRuleFromRows(rows pgx.Rows) (*domain.IPRule, error) {
	var (
		id        uuid.UUID
		ipAddress string
		ruleType  string
		reason    string
		isActive  bool
		createdAt time.Time
		updatedAt time.Time
	)

	err := rows.Scan(&id, &ipAddress, &ruleType, &reason, &isActive, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	_ = isActive

	return domain.ReconstructIPRule(id, createdAt, updatedAt, ipAddress, ruleType, reason, nil), nil
}
