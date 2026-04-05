package postgres

import (
	"context"
	"fmt"
	"strings"

	"gct/internal/context/iam/generic/authz/domain"
	"gct/internal/kernel/consts"
	shared "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/metadata"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuthzReadRepo implements domain.AuthzReadRepository for the CQRS read side.
type AuthzReadRepo struct {
	pool     *pgxpool.Pool
	builder  squirrel.StatementBuilderType
	metadata *metadata.GenericMetadataRepo
}

// NewAuthzReadRepo creates a new AuthzReadRepo.
func NewAuthzReadRepo(pool *pgxpool.Pool) *AuthzReadRepo {
	return &AuthzReadRepo{
		pool:     pool,
		builder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		metadata: metadata.NewGenericMetadataRepo(pool),
	}
}

// GetRole returns a RoleView for the given role ID.
func (r *AuthzReadRepo) GetRole(ctx context.Context, id uuid.UUID) (result *domain.RoleView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "AuthzReadRepo.GetRole")
	defer func() { end(err) }()

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
func (r *AuthzReadRepo) ListRoles(ctx context.Context, pagination shared.Pagination) (items []*domain.RoleView, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "AuthzReadRepo.ListRoles")
	defer func() { end(err) }()

	countSQL, countArgs, err := r.builder.
		Select("COUNT(*)").
		From(consts.TableRole).
		ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

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
func (r *AuthzReadRepo) GetPermission(ctx context.Context, id uuid.UUID) (result *domain.PermissionView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "AuthzReadRepo.GetPermission")
	defer func() { end(err) }()

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
func (r *AuthzReadRepo) ListPermissions(ctx context.Context, pagination shared.Pagination) (items []*domain.PermissionView, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "AuthzReadRepo.ListPermissions")
	defer func() { end(err) }()

	countSQL, countArgs, err := r.builder.
		Select("COUNT(*)").
		From(consts.TablePermission).
		ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

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
func (r *AuthzReadRepo) ListPolicies(ctx context.Context, pagination shared.Pagination) (items []*domain.PolicyView, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "AuthzReadRepo.ListPolicies")
	defer func() { end(err) }()

	countSQL, countArgs, err := r.builder.
		Select("COUNT(*)").
		From(consts.TablePolicy).
		ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, consts.TablePolicy, nil)
	}

	qb := r.builder.
		Select("id", "permission_id", "effect", "priority", "active").
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
		var v domain.PolicyView
		if err := rows.Scan(&v.ID, &v.PermissionID, &v.Effect, &v.Priority, &v.Active); err != nil {
			return nil, 0, apperrors.HandlePgError(err, consts.TablePolicy, nil)
		}
		views = append(views, &v)
	}

	for _, v := range views {
		conds, err := r.metadata.GetAll(ctx, metadata.EntityTypePolicyConditions, v.ID)
		if err != nil {
			return nil, 0, err
		}
		v.Conditions = stringsToConditions(conds)
	}

	return views, total, nil
}

// ListScopes returns a paginated list of ScopeView.
func (r *AuthzReadRepo) ListScopes(ctx context.Context, pagination shared.Pagination) (items []*domain.ScopeView, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "AuthzReadRepo.ListScopes")
	defer func() { end(err) }()

	countSQL, countArgs, err := r.builder.
		Select("COUNT(*)").
		From(consts.TableScope).
		ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

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
// role → permission → scope chain looking for a matching scope entry, then evaluates
// any ABAC policies bound to the matched permissions.
func (r *AuthzReadRepo) CheckAccess(ctx context.Context, roleID uuid.UUID, path, method string, evalCtx domain.EvaluationContext) (result bool, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "AuthzReadRepo.CheckAccess")
	defer func() { end(err) }()

	// Single query: fetch role name together with every scope reachable through
	// the role_permission + permission_scope join chain, plus the permission ID.
	sql, args, err := r.builder.
		Select("r.name", "s.path", "s.method", "rp.permission_id").
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
	matchedPermIDs := map[uuid.UUID]struct{}{}

	for rows.Next() {
		var (
			rName       string
			scopePath   *string
			scopeMethod *string
			permID      *uuid.UUID
		)
		if err := rows.Scan(&rName, &scopePath, &scopeMethod, &permID); err != nil {
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

		if matchScope(*scopePath, *scopeMethod, path, method) && permID != nil {
			matchedPermIDs[*permID] = struct{}{}
		}
	}

	if !foundRole {
		return false, apperrors.HandlePgError(fmt.Errorf("role not found"), consts.TableRole, map[string]any{"id": roleID})
	}

	// No RBAC match — deny.
	if len(matchedPermIDs) == 0 {
		return false, nil
	}

	// Collect permission IDs.
	permIDs := make([]uuid.UUID, 0, len(matchedPermIDs))
	for id := range matchedPermIDs {
		permIDs = append(permIDs, id)
	}

	// Fetch ABAC policies for matched permissions.
	policies, err := r.FindPoliciesByPermissionIDs(ctx, permIDs)
	if err != nil {
		return false, err
	}

	// No policies — RBAC sufficient.
	if len(policies) == 0 {
		return true, nil
	}

	// Inject role_name into evalCtx user attrs.
	if evalCtx.Attrs == nil {
		evalCtx.Attrs = make(map[string]map[string]any)
	}
	if evalCtx.Attrs["user"] == nil {
		evalCtx.Attrs["user"] = make(map[string]any)
	}
	evalCtx.Attrs["user"]["role_name"] = roleName

	evaluator := domain.PolicyEvaluator{}
	effect, matched := evaluator.Evaluate(policies, evalCtx)
	if !matched {
		// No policy matched conditions — RBAC sufficient.
		return true, nil
	}
	return effect == domain.PolicyAllow, nil
}

// FindPoliciesByPermissionIDs returns all policies bound to any of the given permission IDs.
func (r *AuthzReadRepo) FindPoliciesByPermissionIDs(ctx context.Context, permissionIDs []uuid.UUID) (result []*domain.Policy, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "AuthzReadRepo.FindPoliciesByPermissionIDs")
	defer func() { end(err) }()

	if len(permissionIDs) == 0 {
		return nil, nil
	}

	sql, args, err := r.builder.
		Select(policyColumns...).
		From(consts.TablePolicy).
		Where(squirrel.Eq{"permission_id": permissionIDs}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TablePolicy, nil)
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		p, err := scanPolicyFromRows(rows)
		if err != nil {
			return nil, apperrors.HandlePgError(err, consts.TablePolicy, nil)
		}
		policies = append(policies, p)
	}

	for _, p := range policies {
		conds, err := r.metadata.GetAll(ctx, metadata.EntityTypePolicyConditions, p.ID())
		if err != nil {
			return nil, err
		}
		p.SetConditions(stringsToConditions(conds))
	}

	return policies, nil
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
