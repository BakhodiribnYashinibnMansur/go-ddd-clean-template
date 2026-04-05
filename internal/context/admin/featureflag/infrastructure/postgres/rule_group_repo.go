package postgres

import (
	"context"
	"time"

	"gct/internal/context/admin/featureflag/domain"
	"gct/internal/platform/domain/consts"
	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const ruleGroupTable = consts.TableFeatureFlagRuleGroups
const conditionTable = consts.TableFeatureFlagConditions

// RuleGroupWriteRepo implements domain.RuleGroupRepository using PostgreSQL.
type RuleGroupWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewRuleGroupWriteRepo creates a new RuleGroupWriteRepo.
func NewRuleGroupWriteRepo(pool *pgxpool.Pool) *RuleGroupWriteRepo {
	return &RuleGroupWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a rule group and all its conditions.
func (r *RuleGroupWriteRepo) Save(ctx context.Context, rg *domain.RuleGroup) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RuleGroupWriteRepo.Save")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Insert(ruleGroupTable).
		Columns("id", "flag_id", "name", "variation", "priority", "created_at", "updated_at").
		Values(
			rg.ID(),
			rg.FlagID(),
			rg.Name(),
			rg.Variation(),
			rg.Priority(),
			rg.CreatedAt(),
			rg.UpdatedAt(),
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, ruleGroupTable, nil)
	}

	for _, c := range rg.Conditions() {
		if err := r.saveCondition(ctx, c); err != nil {
			return err
		}
	}

	return nil
}

// FindByID retrieves a rule group by ID with its conditions.
func (r *RuleGroupWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (result *domain.RuleGroup, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RuleGroupWriteRepo.FindByID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select("id", "flag_id", "name", "variation", "priority", "created_at", "updated_at").
		From(ruleGroupTable).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var (
		rgID      uuid.UUID
		flagID    uuid.UUID
		name      string
		variation string
		priority  int
		createdAt time.Time
		updatedAt time.Time
	)

	row := r.pool.QueryRow(ctx, sql, args...)
	if err := row.Scan(&rgID, &flagID, &name, &variation, &priority, &createdAt, &updatedAt); err != nil {
		return nil, apperrors.HandlePgError(err, ruleGroupTable, map[string]any{"id": id})
	}

	conditions, err := r.loadConditions(ctx, rgID)
	if err != nil {
		return nil, err
	}

	return domain.ReconstructRuleGroup(rgID, flagID, name, variation, priority, createdAt, updatedAt, conditions), nil
}

// Update updates a rule group's fields and replaces all conditions.
func (r *RuleGroupWriteRepo) Update(ctx context.Context, rg *domain.RuleGroup) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RuleGroupWriteRepo.Update")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Update(ruleGroupTable).
		Set("name", rg.Name()).
		Set("variation", rg.Variation()).
		Set("priority", rg.Priority()).
		Set("updated_at", rg.UpdatedAt()).
		Where(squirrel.Eq{"id": rg.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, ruleGroupTable, nil)
	}

	// Replace conditions: delete old, insert new.
	if err := r.DeleteConditionsByRuleGroupID(ctx, rg.ID()); err != nil {
		return err
	}

	for _, c := range rg.Conditions() {
		if err := r.saveCondition(ctx, c); err != nil {
			return err
		}
	}

	return nil
}

// Delete removes a rule group by ID. FK cascades handle conditions.
func (r *RuleGroupWriteRepo) Delete(ctx context.Context, id uuid.UUID) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RuleGroupWriteRepo.Delete")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Delete(ruleGroupTable).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, ruleGroupTable, nil)
	}

	return nil
}

// FindByFlagID retrieves all rule groups for a flag, ordered by priority.
func (r *RuleGroupWriteRepo) FindByFlagID(ctx context.Context, flagID uuid.UUID) (result []*domain.RuleGroup, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RuleGroupWriteRepo.FindByFlagID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select("id", "flag_id", "name", "variation", "priority", "created_at", "updated_at").
		From(ruleGroupTable).
		Where(squirrel.Eq{"flag_id": flagID}).
		OrderBy("priority ASC").
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, ruleGroupTable, nil)
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
			return nil, apperrors.HandlePgError(err, ruleGroupTable, nil)
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

// SaveCondition inserts a single condition for a rule group.
func (r *RuleGroupWriteRepo) SaveCondition(ctx context.Context, rgID uuid.UUID, c domain.Condition) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RuleGroupWriteRepo.SaveCondition")
	defer func() { end(err) }()

	_ = rgID // condition already carries its rule_group_id
	return r.saveCondition(ctx, c)
}

// DeleteConditionsByRuleGroupID removes all conditions for a rule group.
func (r *RuleGroupWriteRepo) DeleteConditionsByRuleGroupID(ctx context.Context, rgID uuid.UUID) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RuleGroupWriteRepo.DeleteConditionsByRuleGroupID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Delete(conditionTable).
		Where(squirrel.Eq{"rule_group_id": rgID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, conditionTable, nil)
	}

	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (r *RuleGroupWriteRepo) saveCondition(ctx context.Context, c domain.Condition) error {
	sql, args, err := r.builder.
		Insert(conditionTable).
		Columns("id", "rule_group_id", "attribute", "operator", "value").
		Values(c.ID(), c.RuleGroupID(), c.Attribute(), c.Operator(), c.Value()).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, conditionTable, nil)
	}

	return nil
}

func (r *RuleGroupWriteRepo) loadConditions(ctx context.Context, ruleGroupID uuid.UUID) ([]domain.Condition, error) {
	sql, args, err := r.builder.
		Select("id", "rule_group_id", "attribute", "operator", "value").
		From(conditionTable).
		Where(squirrel.Eq{"rule_group_id": ruleGroupID}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, conditionTable, nil)
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
			return nil, apperrors.HandlePgError(err, conditionTable, nil)
		}
		conditions = append(conditions, domain.ReconstructCondition(id, rgID, attr, op, value))
	}

	return conditions, nil
}
