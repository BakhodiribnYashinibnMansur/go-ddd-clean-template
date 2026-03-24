package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"gct/internal/authz/domain"
	"gct/internal/shared/domain/consts"
	shared "gct/internal/shared/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuthzReadRepo implements domain.AuthzReadRepository for the CQRS read side.
type AuthzReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewAuthzReadRepo creates a new AuthzReadRepo.
func NewAuthzReadRepo(pool *pgxpool.Pool) *AuthzReadRepo {
	return &AuthzReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// GetRole returns a RoleView for the given role ID.
func (r *AuthzReadRepo) GetRole(ctx context.Context, id uuid.UUID) (*domain.RoleView, error) {
	sql, args, err := r.builder.
		Select("id", "name", "description").
		From(consts.TableRole).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var view domain.RoleView
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&view.ID, &view.Name, &view.Description)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableRole, map[string]any{"id": id})
	}

	return &view, nil
}

// ListRoles returns a paginated list of RoleView.
func (r *AuthzReadRepo) ListRoles(ctx context.Context, pagination shared.Pagination) ([]*domain.RoleView, int64, error) {
	countSQL, countArgs, err := r.builder.
		Select("COUNT(*)").
		From(consts.TableRole).
		ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var total int64
	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TableRole, nil)
	}

	qb := r.builder.
		Select("id", "name", "description").
		From(consts.TableRole).
		Limit(uint64(pagination.Limit)).
		Offset(uint64(pagination.Offset))

	if pagination.SortBy != "" {
		order := "ASC"
		if pagination.SortOrder == "DESC" {
			order = "DESC"
		}
		qb = qb.OrderBy(fmt.Sprintf("%s %s", pagination.SortBy, order))
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TableRole, nil)
	}
	defer rows.Close()

	var views []*domain.RoleView
	for rows.Next() {
		var v domain.RoleView
		if err := rows.Scan(&v.ID, &v.Name, &v.Description); err != nil {
			return nil, 0, apperrors.HandlePgError(err, consts.TableRole, nil)
		}
		views = append(views, &v)
	}

	return views, total, nil
}

// GetPermission returns a PermissionView for the given permission ID.
func (r *AuthzReadRepo) GetPermission(ctx context.Context, id uuid.UUID) (*domain.PermissionView, error) {
	sql, args, err := r.builder.
		Select("id", "parent_id", "name", "description").
		From(consts.TablePermission).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var view domain.PermissionView
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&view.ID, &view.ParentID, &view.Name, &view.Description)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TablePermission, map[string]any{"id": id})
	}

	return &view, nil
}

// ListPermissions returns a paginated list of PermissionView.
func (r *AuthzReadRepo) ListPermissions(ctx context.Context, pagination shared.Pagination) ([]*domain.PermissionView, int64, error) {
	countSQL, countArgs, err := r.builder.
		Select("COUNT(*)").
		From(consts.TablePermission).
		ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var total int64
	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TablePermission, nil)
	}

	qb := r.builder.
		Select("id", "parent_id", "name", "description").
		From(consts.TablePermission).
		Limit(uint64(pagination.Limit)).
		Offset(uint64(pagination.Offset))

	if pagination.SortBy != "" {
		order := "ASC"
		if pagination.SortOrder == "DESC" {
			order = "DESC"
		}
		qb = qb.OrderBy(fmt.Sprintf("%s %s", pagination.SortBy, order))
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TablePermission, nil)
	}
	defer rows.Close()

	var views []*domain.PermissionView
	for rows.Next() {
		var v domain.PermissionView
		if err := rows.Scan(&v.ID, &v.ParentID, &v.Name, &v.Description); err != nil {
			return nil, 0, apperrors.HandlePgError(err, consts.TablePermission, nil)
		}
		views = append(views, &v)
	}

	return views, total, nil
}

// ListPolicies returns a paginated list of PolicyView.
func (r *AuthzReadRepo) ListPolicies(ctx context.Context, pagination shared.Pagination) ([]*domain.PolicyView, int64, error) {
	countSQL, countArgs, err := r.builder.
		Select("COUNT(*)").
		From(consts.TablePolicy).
		ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var total int64
	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TablePolicy, nil)
	}

	qb := r.builder.
		Select("id", "permission_id", "effect", "priority", "active", "conditions").
		From(consts.TablePolicy).
		Limit(uint64(pagination.Limit)).
		Offset(uint64(pagination.Offset))

	if pagination.SortBy != "" {
		order := "ASC"
		if pagination.SortOrder == "DESC" {
			order = "DESC"
		}
		qb = qb.OrderBy(fmt.Sprintf("%s %s", pagination.SortBy, order))
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TablePolicy, nil)
	}
	defer rows.Close()

	var views []*domain.PolicyView
	for rows.Next() {
		var (
			v        domain.PolicyView
			condJSON []byte
		)
		if err := rows.Scan(&v.ID, &v.PermissionID, &v.Effect, &v.Priority, &v.Active, &condJSON); err != nil {
			return nil, 0, apperrors.HandlePgError(err, consts.TablePolicy, nil)
		}
		if len(condJSON) > 0 {
			_ = json.Unmarshal(condJSON, &v.Conditions)
		}
		views = append(views, &v)
	}

	return views, total, nil
}

// ListScopes returns a paginated list of ScopeView.
func (r *AuthzReadRepo) ListScopes(ctx context.Context, pagination shared.Pagination) ([]*domain.ScopeView, int64, error) {
	countSQL, countArgs, err := r.builder.
		Select("COUNT(*)").
		From(consts.TableScope).
		ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var total int64
	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TableScope, nil)
	}

	qb := r.builder.
		Select("path", "method").
		From(consts.TableScope).
		Limit(uint64(pagination.Limit)).
		Offset(uint64(pagination.Offset))

	if pagination.SortBy != "" {
		order := "ASC"
		if pagination.SortOrder == "DESC" {
			order = "DESC"
		}
		qb = qb.OrderBy(fmt.Sprintf("%s %s", pagination.SortBy, order))
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TableScope, nil)
	}
	defer rows.Close()

	var views []*domain.ScopeView
	for rows.Next() {
		var v domain.ScopeView
		if err := rows.Scan(&v.Path, &v.Method); err != nil {
			return nil, 0, apperrors.HandlePgError(err, consts.TableScope, nil)
		}
		views = append(views, &v)
	}

	return views, total, nil
}
