package postgres

import (
	"context"
	"time"

	"gct/internal/context/admin/featureflag/domain"
	"gct/internal/kernel/consts"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var readColumns = []string{
	"id", "key", "name", "flag_type", "default_value",
	"description", "rollout_percentage", "is_active", "created_at", "updated_at",
}

// FeatureFlagReadRepo implements domain.FeatureFlagReadRepository for the CQRS read side.
type FeatureFlagReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewFeatureFlagReadRepo creates a new FeatureFlagReadRepo.
func NewFeatureFlagReadRepo(pool *pgxpool.Pool) *FeatureFlagReadRepo {
	return &FeatureFlagReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// FindByID returns a FeatureFlagView for the given ID.
func (r *FeatureFlagReadRepo) FindByID(ctx context.Context, id uuid.UUID) (result *domain.FeatureFlagView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "FeatureFlagReadRepo.FindByID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select(readColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id, "deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var (
		ffID              uuid.UUID
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

	row := r.pool.QueryRow(ctx, sql, args...)
	if err := row.Scan(&ffID, &key, &name, &flagType, &defaultValue, &description, &rolloutPercentage, &isActive, &createdAt, &updatedAt); err != nil {
		return nil, apperrors.HandlePgError(err, tableName, map[string]any{"id": id})
	}

	ruleGroupViews, err := r.loadRuleGroupViews(ctx, ffID)
	if err != nil {
		return nil, err
	}

	return &domain.FeatureFlagView{
		ID:                ffID,
		Name:              name,
		Key:               key,
		Description:       description,
		FlagType:          flagType,
		DefaultValue:      defaultValue,
		RolloutPercentage: rolloutPercentage,
		IsActive:          isActive,
		RuleGroups:        ruleGroupViews,
		CreatedAt:         createdAt.Format(time.RFC3339),
		UpdatedAt:         updatedAt.Format(time.RFC3339),
	}, nil
}

// List returns a paginated list of FeatureFlagView with optional filters.
func (r *FeatureFlagReadRepo) List(ctx context.Context, filter domain.FeatureFlagFilter) (items []*domain.FeatureFlagView, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "FeatureFlagReadRepo.List")
	defer func() { end(err) }()

	conds := squirrel.And{squirrel.Eq{"deleted_at": nil}}
	if filter.Search != nil {
		conds = append(conds, squirrel.ILike{"name": "%" + *filter.Search + "%"})
	}
	if filter.Enabled != nil {
		conds = append(conds, squirrel.Eq{"is_active": *filter.Enabled})
	}

	// Count total.
	countQB := r.builder.Select("COUNT(*)").From(tableName).Where(conds)
	countSQL, countArgs, err := countQB.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}

	// Fetch page.
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	qb := r.builder.
		Select(readColumns...).
		From(tableName).
		Where(conds).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(filter.Offset))

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}
	defer rows.Close()

	var views []*domain.FeatureFlagView
	for rows.Next() {
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

		if err := rows.Scan(&id, &key, &name, &flagType, &defaultValue, &description, &rolloutPercentage, &isActive, &createdAt, &updatedAt); err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}

		ruleGroupViews, err := r.loadRuleGroupViews(ctx, id)
		if err != nil {
			return nil, 0, err
		}

		views = append(views, &domain.FeatureFlagView{
			ID:                id,
			Name:              name,
			Key:               key,
			Description:       description,
			FlagType:          flagType,
			DefaultValue:      defaultValue,
			RolloutPercentage: rolloutPercentage,
			IsActive:          isActive,
			RuleGroups:        ruleGroupViews,
			CreatedAt:         createdAt.Format(time.RFC3339),
			UpdatedAt:         updatedAt.Format(time.RFC3339),
		})
	}

	return views, total, nil
}

// ---------------------------------------------------------------------------
// View loaders
// ---------------------------------------------------------------------------

func (r *FeatureFlagReadRepo) loadRuleGroupViews(ctx context.Context, flagID uuid.UUID) ([]domain.RuleGroupView, error) {
	sql, args, err := r.builder.
		Select("id", "name", "variation", "priority", "created_at", "updated_at").
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

	var views []domain.RuleGroupView
	for rows.Next() {
		var (
			id        uuid.UUID
			name      string
			variation string
			priority  int
			createdAt time.Time
			updatedAt time.Time
		)
		if err := rows.Scan(&id, &name, &variation, &priority, &createdAt, &updatedAt); err != nil {
			return nil, apperrors.HandlePgError(err, consts.TableFeatureFlagRuleGroups, nil)
		}

		condViews, err := r.loadConditionViews(ctx, id)
		if err != nil {
			return nil, err
		}

		views = append(views, domain.RuleGroupView{
			ID:         id,
			Name:       name,
			Variation:  variation,
			Priority:   priority,
			Conditions: condViews,
			CreatedAt:  createdAt.Format(time.RFC3339),
			UpdatedAt:  updatedAt.Format(time.RFC3339),
		})
	}

	return views, nil
}

func (r *FeatureFlagReadRepo) loadConditionViews(ctx context.Context, ruleGroupID uuid.UUID) ([]domain.ConditionView, error) {
	sql, args, err := r.builder.
		Select("id", "attribute", "operator", "value").
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

	var views []domain.ConditionView
	for rows.Next() {
		var (
			id    uuid.UUID
			attr  string
			op    string
			value string
		)
		if err := rows.Scan(&id, &attr, &op, &value); err != nil {
			return nil, apperrors.HandlePgError(err, consts.TableFeatureFlagConditions, nil)
		}
		views = append(views, domain.ConditionView{
			ID:        id,
			Attribute: attr,
			Operator:  op,
			Value:     value,
		})
	}

	return views, nil
}
