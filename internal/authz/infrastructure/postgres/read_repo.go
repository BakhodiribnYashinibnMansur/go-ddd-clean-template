package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

// CheckAccess returns true if the role identified by roleID is allowed to access the given path+method.
// Super-admin roles bypass all checks. For other roles the method walks the
// role → permission → scope chain looking for a matching scope entry.
func (r *AuthzReadRepo) CheckAccess(ctx context.Context, roleID uuid.UUID, path, method string) (bool, error) {
	// Single query: fetch role name together with every scope reachable through
	// the role_permission + permission_scope join chain.
	sql, args, err := r.builder.
		Select("r.name", "s.path", "s.method").
		From(consts.TableRole + " r").
		LeftJoin(rolePermissionTable + " rp ON r.id = rp.role_id").
		LeftJoin(permissionScopeTable + " ps ON rp.permission_id = ps.permission_id").
		LeftJoin(consts.TableScope + " s ON ps.path = s.path AND ps.method = s.method").
		Where(squirrel.Eq{"r.id": roleID}).
		ToSql()
	if err != nil {
		return false, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return false, apperrors.HandlePgError(err, consts.TableRole, map[string]any{"id": roleID})
	}
	defer rows.Close()

	var roleName string
	foundRole := false

	for rows.Next() {
		var (
			rName       string
			scopePath   *string
			scopeMethod *string
		)
		if err := rows.Scan(&rName, &scopePath, &scopeMethod); err != nil {
			return false, apperrors.HandlePgError(err, consts.TableRole, map[string]any{"id": roleID})
		}

		if !foundRole {
			roleName = rName
			foundRole = true

			// Super-admin bypass — grant access unconditionally.
			if strings.ToLower(roleName) == "super_admin" {
				return true, nil
			}
		}

		// LEFT JOIN may produce NULL scope columns when no scopes are assigned.
		if scopePath == nil || scopeMethod == nil {
			continue
		}

		if matchScope(*scopePath, *scopeMethod, path, method) {
			return true, nil
		}
	}

	if !foundRole {
		return false, apperrors.HandlePgError(fmt.Errorf("role not found"), consts.TableRole, map[string]any{"id": roleID})
	}

	return false, nil
}

// matchScope checks whether a scope entry matches the requested path and method.
// It supports exact match, wildcard ("*") method, and prefix matching when the
// scope path ends with "*" (e.g. "/api/v1/users*" matches "/api/v1/users/123").
func matchScope(scopePath, scopeMethod, requestPath, requestMethod string) bool {
	// Method must match exactly or be a wildcard.
	if scopeMethod != "*" && !strings.EqualFold(scopeMethod, requestMethod) {
		return false
	}

	// Exact path match.
	if scopePath == requestPath {
		return true
	}

	// Wildcard path — matches everything.
	if scopePath == "*" {
		return true
	}

	// Prefix match: scope path ending with "*".
	if strings.HasSuffix(scopePath, "*") {
		prefix := strings.TrimSuffix(scopePath, "*")
		if strings.HasPrefix(requestPath, prefix) {
			return true
		}
	}

	return false
}
