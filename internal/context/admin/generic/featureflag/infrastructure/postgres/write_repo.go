package postgres

import (
	"context"
	"time"

	"gct/internal/context/admin/generic/featureflag/domain"
	"gct/internal/kernel/consts"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableFeatureFlags

var writeColumns = []string{
	"id", "key", "name", "flag_type", "value", "default_value",
	"description", "rollout_percentage", "is_active", "created_at", "updated_at",
}

var selectColumns = []string{
	"id", "key", "name", "flag_type", "default_value",
	"description", "rollout_percentage", "is_active", "created_at", "updated_at",
}

// FeatureFlagWriteRepo implements domain.FeatureFlagRepository using PostgreSQL.
type FeatureFlagWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewFeatureFlagWriteRepo creates a new FeatureFlagWriteRepo.
func NewFeatureFlagWriteRepo(pool *pgxpool.Pool) *FeatureFlagWriteRepo {
	return &FeatureFlagWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new FeatureFlag aggregate into the database.
func (r *FeatureFlagWriteRepo) Save(ctx context.Context, ff *domain.FeatureFlag) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "FeatureFlagWriteRepo.Save")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			ff.ID(),
			ff.Key(),
			ff.Name(),
			ff.FlagType(),
			ff.DefaultValue(),
			ff.DefaultValue(),
			ff.Description(),
			ff.RolloutPercentage(),
			ff.IsActive(),
			ff.CreatedAt(),
			ff.UpdatedAt(),
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

// FindByID retrieves a FeatureFlag aggregate by ID.
func (r *FeatureFlagWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (result *domain.FeatureFlag, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "FeatureFlagWriteRepo.FindByID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select(selectColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id, "deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	ff, err := scanFeatureFlag(row)
	if err != nil {
		return nil, err
	}

	ruleGroups, err := r.loadRuleGroups(ctx, ff.ID())
	if err != nil {
		return nil, err
	}
	for _, rg := range ruleGroups {
		ff.AddRuleGroup(rg)
	}

	return ff, nil
}

// FindByKey retrieves a FeatureFlag aggregate by its unique key.
func (r *FeatureFlagWriteRepo) FindByKey(ctx context.Context, key string) (result *domain.FeatureFlag, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "FeatureFlagWriteRepo.FindByKey")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select(selectColumns...).
		From(tableName).
		Where(squirrel.Eq{"key": key, "deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	ff, err := scanFeatureFlag(row)
	if err != nil {
		return nil, err
	}

	ruleGroups, err := r.loadRuleGroups(ctx, ff.ID())
	if err != nil {
		return nil, err
	}
	for _, rg := range ruleGroups {
		ff.AddRuleGroup(rg)
	}

	return ff, nil
}

// FindAll retrieves all non-deleted FeatureFlag aggregates.
func (r *FeatureFlagWriteRepo) FindAll(ctx context.Context) (result []*domain.FeatureFlag, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "FeatureFlagWriteRepo.FindAll")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select(selectColumns...).
		From(tableName).
		Where(squirrel.Eq{"deleted_at": nil}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}
	defer rows.Close()

	var flags []*domain.FeatureFlag
	for rows.Next() {
		ff, err := scanFeatureFlagFromRows(rows)
		if err != nil {
			return nil, apperrors.HandlePgError(err, tableName, nil)
		}
		flags = append(flags, ff)
	}

	for _, ff := range flags {
		ruleGroups, err := r.loadRuleGroups(ctx, ff.ID())
		if err != nil {
			return nil, err
		}
		for _, rg := range ruleGroups {
			ff.AddRuleGroup(rg)
		}
	}

	return flags, nil
}

// Update updates a FeatureFlag aggregate in the database.
func (r *FeatureFlagWriteRepo) Update(ctx context.Context, ff *domain.FeatureFlag) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "FeatureFlagWriteRepo.Update")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Update(tableName).
		Set("name", ff.Name()).
		Set("key", ff.Key()).
		Set("flag_type", ff.FlagType()).
		Set("default_value", ff.DefaultValue()).
		Set("value", ff.DefaultValue()).
		Set("description", ff.Description()).
		Set("rollout_percentage", ff.RolloutPercentage()).
		Set("is_active", ff.IsActive()).
		Set("updated_at", ff.UpdatedAt()).
		Where(squirrel.Eq{"id": ff.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// Delete removes a FeatureFlag by ID.
func (r *FeatureFlagWriteRepo) Delete(ctx context.Context, id uuid.UUID) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "FeatureFlagWriteRepo.Delete")
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

// ---------------------------------------------------------------------------
// Rule group / condition loaders
// ---------------------------------------------------------------------------

func (r *FeatureFlagWriteRepo) loadRuleGroups(ctx context.Context, flagID uuid.UUID) ([]*domain.RuleGroup, error) {
	sql, args, err := r.builder.
		Select("id", "flag_id", "name", "variation", "priority", "created_at", "updated_at").
		From(consts.TableFeatureFlagRuleGroups).
		Where(squirrel.Eq{"flag_id": flagID}).
		OrderBy("priority ASC").
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableFeatureFlagRuleGroups, nil)
	}
	defer rows.Close()

	var ruleGroups []*domain.RuleGroup
	for rows.Next() {
		var (
			id        uuid.UUID
			fID       uuid.UUID
			name      string
			variation string
			priority  int
			createdAt time.Time
			updatedAt time.Time
		)
		if err := rows.Scan(&id, &fID, &name, &variation, &priority, &createdAt, &updatedAt); err != nil {
			return nil, apperrors.HandlePgError(err, consts.TableFeatureFlagRuleGroups, nil)
		}

		conditions, err := r.loadConditions(ctx, id)
		if err != nil {
			return nil, err
		}

		rg := domain.ReconstructRuleGroup(id, fID, name, variation, priority, createdAt, updatedAt, conditions)
		ruleGroups = append(ruleGroups, rg)
	}

	return ruleGroups, nil
}

func (r *FeatureFlagWriteRepo) loadConditions(ctx context.Context, ruleGroupID uuid.UUID) ([]domain.Condition, error) {
	sql, args, err := r.builder.
		Select("id", "rule_group_id", "attribute", "operator", "value").
		From(consts.TableFeatureFlagConditions).
		Where(squirrel.Eq{"rule_group_id": ruleGroupID}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableFeatureFlagConditions, nil)
	}
	defer rows.Close()

	var conditions []domain.Condition
	for rows.Next() {
		var (
			id    uuid.UUID
			rgID  uuid.UUID
			attr  string
			op    string
			value string
		)
		if err := rows.Scan(&id, &rgID, &attr, &op, &value); err != nil {
			return nil, apperrors.HandlePgError(err, consts.TableFeatureFlagConditions, nil)
		}
		conditions = append(conditions, domain.ReconstructCondition(id, rgID, attr, op, value))
	}

	return conditions, nil
}

// ---------------------------------------------------------------------------
// Scanners
// ---------------------------------------------------------------------------

func scanFeatureFlag(row pgx.Row) (*domain.FeatureFlag, error) {
	var (
		id                uuid.UUID
		key               string
		name              string
		flagType          string
		defaultValue      string
		description       string
		rolloutPercentage int
		isActive          bool
		createdAt         time.Time
		updatedAt         time.Time
	)

	err := row.Scan(&id, &key, &name, &flagType, &defaultValue, &description, &rolloutPercentage, &isActive, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, map[string]any{"id": id})
	}

	return domain.ReconstructFeatureFlag(id, createdAt, updatedAt, nil, name, key, description, flagType, defaultValue, rolloutPercentage, isActive, nil), nil
}

func scanFeatureFlagFromRows(rows pgx.Rows) (*domain.FeatureFlag, error) {
	var (
		id                uuid.UUID
		key               string
		name              string
		flagType          string
		defaultValue      string
		description       string
		rolloutPercentage int
		isActive          bool
		createdAt         time.Time
		updatedAt         time.Time
	)

	err := rows.Scan(&id, &key, &name, &flagType, &defaultValue, &description, &rolloutPercentage, &isActive, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return domain.ReconstructFeatureFlag(id, createdAt, updatedAt, nil, name, key, description, flagType, defaultValue, rolloutPercentage, isActive, nil), nil
}
