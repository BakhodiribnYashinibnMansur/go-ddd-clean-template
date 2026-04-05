package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gct/internal/context/iam/generic/authz/domain"
	"gct/internal/kernel/consts"
	shared "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/metadata"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	roleTable            = consts.TableRole
	permissionTable      = consts.TablePermission
	policyTable          = consts.TablePolicy
	scopeTable           = consts.TableScope
	rolePermissionTable  = "role_permission"
	permissionScopeTable = "permission_scope"
)

var (
	roleColumns       = []string{"id", "name", "description", "created_at", "updated_at"}
	permissionColumns = []string{"id", "parent_id", "name", "description", "created_at", "updated_at"}
	policyColumns     = []string{"id", "permission_id", "effect", "priority", "active", "created_at", "updated_at"}
	scopeColumns      = []string{"path", "method"}
)

// ---------------------------------------------------------------------------
// RoleWriteRepo
// ---------------------------------------------------------------------------

// RoleWriteRepo implements domain.RoleRepository using PostgreSQL.
type RoleWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewRoleWriteRepo creates a new RoleWriteRepo.
func NewRoleWriteRepo(pool *pgxpool.Pool) *RoleWriteRepo {
	return &RoleWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new role.
func (r *RoleWriteRepo) Save(ctx context.Context, role *domain.Role) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RoleWriteRepo.Save")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Insert(roleTable).
		Columns(roleColumns...).
		Values(
			role.ID(),
			role.Name(),
			role.Description(),
			role.CreatedAt(),
			role.UpdatedAt(),
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, roleTable, nil)
	}
	return nil
}

// FindByID retrieves a role by ID.
func (r *RoleWriteRepo) FindByID(ctx context.Context, id domain.RoleID) (result *domain.Role, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RoleWriteRepo.FindByID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select(roleColumns...).
		From(roleTable).
		Where(squirrel.Eq{"id": id.UUID()}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanRole(row)
}

// Update updates an existing role.
func (r *RoleWriteRepo) Update(ctx context.Context, role *domain.Role) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RoleWriteRepo.Update")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Update(roleTable).
		Set("name", role.Name()).
		Set("description", role.Description()).
		Set("updated_at", role.UpdatedAt()).
		Where(squirrel.Eq{"id": role.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, roleTable, nil)
	}
	return nil
}

// Delete deletes a role by ID.
func (r *RoleWriteRepo) Delete(ctx context.Context, id domain.RoleID) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RoleWriteRepo.Delete")
	defer func() { end(err) }()

	return pgxutil.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		// Delete role_permission entries first.
		delRP, argsRP, err := r.builder.
			Delete(rolePermissionTable).
			Where(squirrel.Eq{"role_id": id}).
			ToSql()
		if err != nil {
			return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
		}
		if _, err = tx.Exec(ctx, delRP, argsRP...); err != nil {
			return apperrors.HandlePgError(err, rolePermissionTable, nil)
		}

		// Delete role.
		delSQL, delArgs, err := r.builder.
			Delete(roleTable).
			Where(squirrel.Eq{"id": id.UUID()}).
			ToSql()
		if err != nil {
			return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
		}
		if _, err = tx.Exec(ctx, delSQL, delArgs...); err != nil {
			return apperrors.HandlePgError(err, roleTable, nil)
		}
		return nil
	})
}

// List returns a paginated list of roles.
func (r *RoleWriteRepo) List(ctx context.Context, pagination shared.Pagination) (items []*domain.Role, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RoleWriteRepo.List")
	defer func() { end(err) }()

	countSQL, countArgs, err := r.builder.
		Select("COUNT(*)").
		From(roleTable).
		ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, roleTable, nil)
	}

	qb := r.builder.
		Select(roleColumns...).
		From(roleTable).
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
		return nil, 0, apperrors.HandlePgError(err, roleTable, nil)
	}
	defer rows.Close()

	var roles []*domain.Role
	for rows.Next() {
		role, err := scanRoleFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, roleTable, nil)
		}
		roles = append(roles, role)
	}

	return roles, total, nil
}

// ---------------------------------------------------------------------------
// PermissionWriteRepo
// ---------------------------------------------------------------------------

// PermissionWriteRepo implements domain.PermissionRepository using PostgreSQL.
type PermissionWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewPermissionWriteRepo creates a new PermissionWriteRepo.
func NewPermissionWriteRepo(pool *pgxpool.Pool) *PermissionWriteRepo {
	return &PermissionWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new permission.
func (r *PermissionWriteRepo) Save(ctx context.Context, perm *domain.Permission) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "PermissionWriteRepo.Save")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Insert(permissionTable).
		Columns(permissionColumns...).
		Values(
			perm.ID(),
			perm.ParentID(),
			perm.Name(),
			perm.Description(),
			perm.CreatedAt(),
			perm.UpdatedAt(),
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, permissionTable, nil)
	}
	return nil
}

// FindByID retrieves a permission by ID.
func (r *PermissionWriteRepo) FindByID(ctx context.Context, id domain.PermissionID) (result *domain.Permission, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "PermissionWriteRepo.FindByID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select(permissionColumns...).
		From(permissionTable).
		Where(squirrel.Eq{"id": id.UUID()}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanPermission(row)
}

// Update updates an existing permission.
func (r *PermissionWriteRepo) Update(ctx context.Context, perm *domain.Permission) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "PermissionWriteRepo.Update")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Update(permissionTable).
		Set("parent_id", perm.ParentID()).
		Set("name", perm.Name()).
		Set("description", perm.Description()).
		Set("updated_at", perm.UpdatedAt()).
		Where(squirrel.Eq{"id": perm.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, permissionTable, nil)
	}
	return nil
}

// Delete deletes a permission by ID.
func (r *PermissionWriteRepo) Delete(ctx context.Context, id domain.PermissionID) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "PermissionWriteRepo.Delete")
	defer func() { end(err) }()

	return pgxutil.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		// Delete permission_scope entries first.
		delPS, argsPS, err := r.builder.
			Delete(permissionScopeTable).
			Where(squirrel.Eq{"permission_id": id}).
			ToSql()
		if err != nil {
			return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
		}
		if _, err = tx.Exec(ctx, delPS, argsPS...); err != nil {
			return apperrors.HandlePgError(err, permissionScopeTable, nil)
		}

		// Delete role_permission entries.
		delRP, argsRP, err := r.builder.
			Delete(rolePermissionTable).
			Where(squirrel.Eq{"permission_id": id}).
			ToSql()
		if err != nil {
			return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
		}
		if _, err = tx.Exec(ctx, delRP, argsRP...); err != nil {
			return apperrors.HandlePgError(err, rolePermissionTable, nil)
		}

		// Delete permission.
		delSQL, delArgs, err := r.builder.
			Delete(permissionTable).
			Where(squirrel.Eq{"id": id.UUID()}).
			ToSql()
		if err != nil {
			return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
		}
		if _, err = tx.Exec(ctx, delSQL, delArgs...); err != nil {
			return apperrors.HandlePgError(err, permissionTable, nil)
		}
		return nil
	})
}

// List returns a paginated list of permissions.
func (r *PermissionWriteRepo) List(ctx context.Context, pagination shared.Pagination) (items []*domain.Permission, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "PermissionWriteRepo.List")
	defer func() { end(err) }()

	countSQL, countArgs, err := r.builder.
		Select("COUNT(*)").
		From(permissionTable).
		ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, permissionTable, nil)
	}

	qb := r.builder.
		Select(permissionColumns...).
		From(permissionTable).
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
		return nil, 0, apperrors.HandlePgError(err, permissionTable, nil)
	}
	defer rows.Close()

	var perms []*domain.Permission
	for rows.Next() {
		p, err := scanPermissionFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, permissionTable, nil)
		}
		perms = append(perms, p)
	}

	return perms, total, nil
}

// ---------------------------------------------------------------------------
// PolicyWriteRepo
// ---------------------------------------------------------------------------

// PolicyWriteRepo implements domain.PolicyRepository using PostgreSQL.
type PolicyWriteRepo struct {
	pool     *pgxpool.Pool
	builder  squirrel.StatementBuilderType
	metadata *metadata.GenericMetadataRepo
}

// NewPolicyWriteRepo creates a new PolicyWriteRepo.
func NewPolicyWriteRepo(pool *pgxpool.Pool) *PolicyWriteRepo {
	return &PolicyWriteRepo{
		pool:     pool,
		builder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		metadata: metadata.NewGenericMetadataRepo(pool),
	}
}

// Save inserts a new policy.
func (r *PolicyWriteRepo) Save(ctx context.Context, policy *domain.Policy) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "PolicyWriteRepo.Save")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Insert(policyTable).
		Columns(policyColumns...).
		Values(
			policy.ID(),
			policy.PermissionID(),
			string(policy.Effect()),
			policy.Priority(),
			policy.IsActive(),
			policy.CreatedAt(),
			policy.UpdatedAt(),
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, policyTable, nil)
	}

	if err := r.metadata.SetMany(ctx, metadata.EntityTypePolicyConditions, policy.ID(), conditionsToStrings(policy.Conditions())); err != nil {
		return err
	}

	return nil
}

// FindByID retrieves a policy by ID.
func (r *PolicyWriteRepo) FindByID(ctx context.Context, id domain.PolicyID) (result *domain.Policy, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "PolicyWriteRepo.FindByID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select(policyColumns...).
		From(policyTable).
		Where(squirrel.Eq{"id": id.UUID()}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	policy, err := scanPolicy(row)
	if err != nil {
		return nil, err
	}

	conds, err := r.metadata.GetAll(ctx, metadata.EntityTypePolicyConditions, policy.ID())
	if err != nil {
		return nil, err
	}
	policy.SetConditions(stringsToConditions(conds))

	return policy, nil
}

// Update updates an existing policy.
func (r *PolicyWriteRepo) Update(ctx context.Context, policy *domain.Policy) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "PolicyWriteRepo.Update")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Update(policyTable).
		Set("effect", string(policy.Effect())).
		Set("priority", policy.Priority()).
		Set("active", policy.IsActive()).
		Set("updated_at", policy.UpdatedAt()).
		Where(squirrel.Eq{"id": policy.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, policyTable, nil)
	}

	if err := r.metadata.SetMany(ctx, metadata.EntityTypePolicyConditions, policy.ID(), conditionsToStrings(policy.Conditions())); err != nil {
		return err
	}

	return nil
}

// Delete deletes a policy by ID.
func (r *PolicyWriteRepo) Delete(ctx context.Context, id domain.PolicyID) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "PolicyWriteRepo.Delete")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Delete(policyTable).
		Where(squirrel.Eq{"id": id.UUID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, policyTable, nil)
	}

	if err := r.metadata.DeleteAll(ctx, metadata.EntityTypePolicyConditions, id.UUID()); err != nil {
		return err
	}

	return nil
}

// List returns a paginated list of policies.
func (r *PolicyWriteRepo) List(ctx context.Context, pagination shared.Pagination) (items []*domain.Policy, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "PolicyWriteRepo.List")
	defer func() { end(err) }()

	countSQL, countArgs, err := r.builder.
		Select("COUNT(*)").
		From(policyTable).
		ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, policyTable, nil)
	}

	qb := r.builder.
		Select(policyColumns...).
		From(policyTable).
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
		return nil, 0, apperrors.HandlePgError(err, policyTable, nil)
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		p, err := scanPolicyFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, policyTable, nil)
		}
		policies = append(policies, p)
	}

	for _, p := range policies {
		conds, err := r.metadata.GetAll(ctx, metadata.EntityTypePolicyConditions, p.ID())
		if err != nil {
			return nil, 0, err
		}
		p.SetConditions(stringsToConditions(conds))
	}

	return policies, total, nil
}

// FindByPermissionID returns all policies for a given permission ID.
func (r *PolicyWriteRepo) FindByPermissionID(ctx context.Context, permissionID domain.PermissionID) (result []*domain.Policy, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "PolicyWriteRepo.FindByPermissionID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select(policyColumns...).
		From(policyTable).
		Where(squirrel.Eq{"permission_id": permissionID.UUID()}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, policyTable, nil)
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		p, err := scanPolicyFromRows(rows)
		if err != nil {
			return nil, apperrors.HandlePgError(err, policyTable, nil)
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

// ---------------------------------------------------------------------------
// ScopeWriteRepo
// ---------------------------------------------------------------------------

// ScopeWriteRepo implements domain.ScopeRepository using PostgreSQL.
type ScopeWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewScopeWriteRepo creates a new ScopeWriteRepo.
func NewScopeWriteRepo(pool *pgxpool.Pool) *ScopeWriteRepo {
	return &ScopeWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new scope.
func (r *ScopeWriteRepo) Save(ctx context.Context, scope domain.Scope) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "ScopeWriteRepo.Save")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Insert(scopeTable).
		Columns(scopeColumns...).
		Values(scope.Path, scope.Method).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, scopeTable, nil)
	}
	return nil
}

// Delete deletes a scope by path and method.
func (r *ScopeWriteRepo) Delete(ctx context.Context, path, method string) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "ScopeWriteRepo.Delete")
	defer func() { end(err) }()

	return pgxutil.WithTx(ctx, r.pool, func(tx pgx.Tx) error {
		// Delete permission_scope entries first.
		delPS, argsPS, err := r.builder.
			Delete(permissionScopeTable).
			Where(squirrel.Eq{"path": path, "method": method}).
			ToSql()
		if err != nil {
			return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
		}
		if _, err = tx.Exec(ctx, delPS, argsPS...); err != nil {
			return apperrors.HandlePgError(err, permissionScopeTable, nil)
		}

		// Delete scope.
		delSQL, delArgs, err := r.builder.
			Delete(scopeTable).
			Where(squirrel.Eq{"path": path, "method": method}).
			ToSql()
		if err != nil {
			return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
		}
		if _, err = tx.Exec(ctx, delSQL, delArgs...); err != nil {
			return apperrors.HandlePgError(err, scopeTable, nil)
		}
		return nil
	})
}

// List returns a paginated list of scopes.
func (r *ScopeWriteRepo) List(ctx context.Context, pagination shared.Pagination) (items []domain.Scope, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "ScopeWriteRepo.List")
	defer func() { end(err) }()

	countSQL, countArgs, err := r.builder.
		Select("COUNT(*)").
		From(scopeTable).
		ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, scopeTable, nil)
	}

	qb := r.builder.
		Select(scopeColumns...).
		From(scopeTable).
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
		return nil, 0, apperrors.HandlePgError(err, scopeTable, nil)
	}
	defer rows.Close()

	var scopes []domain.Scope
	for rows.Next() {
		var s domain.Scope
		if err := rows.Scan(&s.Path, &s.Method); err != nil {
			return nil, 0, apperrors.HandlePgError(err, scopeTable, nil)
		}
		scopes = append(scopes, s)
	}

	return scopes, total, nil
}

// ---------------------------------------------------------------------------
// RolePermissionRepo
// ---------------------------------------------------------------------------

// RolePermissionRepo implements domain.RolePermissionRepository.
type RolePermissionRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewRolePermissionRepo creates a new RolePermissionRepo.
func NewRolePermissionRepo(pool *pgxpool.Pool) *RolePermissionRepo {
	return &RolePermissionRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Assign inserts a role-permission association.
func (r *RolePermissionRepo) Assign(ctx context.Context, roleID domain.RoleID, permissionID domain.PermissionID) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RolePermissionRepo.Assign")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Insert(rolePermissionTable).
		Columns("role_id", "permission_id").
		Values(roleID, permissionID).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, rolePermissionTable, nil)
	}
	return nil
}

// Revoke removes a role-permission association.
func (r *RolePermissionRepo) Revoke(ctx context.Context, roleID domain.RoleID, permissionID domain.PermissionID) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RolePermissionRepo.Revoke")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Delete(rolePermissionTable).
		Where(squirrel.Eq{"role_id": roleID, "permission_id": permissionID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, rolePermissionTable, nil)
	}
	return nil
}

// ---------------------------------------------------------------------------
// PermissionScopeRepo
// ---------------------------------------------------------------------------

// PermissionScopeRepo implements domain.PermissionScopeRepository.
type PermissionScopeRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewPermissionScopeRepo creates a new PermissionScopeRepo.
func NewPermissionScopeRepo(pool *pgxpool.Pool) *PermissionScopeRepo {
	return &PermissionScopeRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Assign inserts a permission-scope association.
func (r *PermissionScopeRepo) Assign(ctx context.Context, permissionID domain.PermissionID, path, method string) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "PermissionScopeRepo.Assign")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Insert(permissionScopeTable).
		Columns("permission_id", "path", "method").
		Values(permissionID, path, method).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, permissionScopeTable, nil)
	}
	return nil
}

// Revoke removes a permission-scope association.
func (r *PermissionScopeRepo) Revoke(ctx context.Context, permissionID domain.PermissionID, path, method string) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "PermissionScopeRepo.Revoke")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Delete(permissionScopeTable).
		Where(squirrel.Eq{"permission_id": permissionID, "path": path, "method": method}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, permissionScopeTable, nil)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Scan helpers
// ---------------------------------------------------------------------------

func scanRole(row pgx.Row) (*domain.Role, error) {
	var (
		id          uuid.UUID
		name        string
		description *string
		ct, ut      interface{}
	)

	err := row.Scan(&id, &name, &description, &ct, &ut)
	if err != nil {
		return nil, apperrors.HandlePgError(err, roleTable, nil)
	}

	return domain.ReconstructRole(id, toTime(ct), toTime(ut), nil, name, description, nil), nil
}

func scanRoleFromRows(rows pgx.Rows) (*domain.Role, error) {
	var (
		id          uuid.UUID
		name        string
		description *string
		ct, ut      interface{}
	)

	err := rows.Scan(&id, &name, &description, &ct, &ut)
	if err != nil {
		return nil, apperrors.HandlePgError(err, roleTable, nil)
	}

	return domain.ReconstructRole(id, toTime(ct), toTime(ut), nil, name, description, nil), nil
}

func scanPermission(row pgx.Row) (*domain.Permission, error) {
	var (
		id          uuid.UUID
		parentID    *uuid.UUID
		name        string
		description *string
		ct, ut      interface{}
	)

	err := row.Scan(&id, &parentID, &name, &description, &ct, &ut)
	if err != nil {
		return nil, apperrors.HandlePgError(err, permissionTable, nil)
	}

	return domain.ReconstructPermission(id, toTime(ct), toTime(ut), nil, parentID, name, description, nil), nil
}

func scanPermissionFromRows(rows pgx.Rows) (*domain.Permission, error) {
	var (
		id          uuid.UUID
		parentID    *uuid.UUID
		name        string
		description *string
		ct, ut      interface{}
	)

	err := rows.Scan(&id, &parentID, &name, &description, &ct, &ut)
	if err != nil {
		return nil, apperrors.HandlePgError(err, permissionTable, nil)
	}

	return domain.ReconstructPermission(id, toTime(ct), toTime(ut), nil, parentID, name, description, nil), nil
}

func scanPolicy(row pgx.Row) (*domain.Policy, error) {
	var (
		id           uuid.UUID
		permissionID uuid.UUID
		effect       string
		priority     int
		active       bool
		ct, ut       interface{}
	)

	err := row.Scan(&id, &permissionID, &effect, &priority, &active, &ct, &ut)
	if err != nil {
		return nil, apperrors.HandlePgError(err, policyTable, nil)
	}

	return domain.ReconstructPolicy(id, toTime(ct), toTime(ut), nil, permissionID, domain.PolicyEffect(effect), priority, active, nil), nil
}

func scanPolicyFromRows(rows pgx.Rows) (*domain.Policy, error) {
	var (
		id           uuid.UUID
		permissionID uuid.UUID
		effect       string
		priority     int
		active       bool
		ct, ut       interface{}
	)

	err := rows.Scan(&id, &permissionID, &effect, &priority, &active, &ct, &ut)
	if err != nil {
		return nil, apperrors.HandlePgError(err, policyTable, nil)
	}

	return domain.ReconstructPolicy(id, toTime(ct), toTime(ut), nil, permissionID, domain.PolicyEffect(effect), priority, active, nil), nil
}

// toTime converts an interface{} (from pgx scan) to time.Time.
func toTime(v interface{}) (t time.Time) {
	if v == nil {
		return t
	}
	switch val := v.(type) {
	case time.Time:
		return val
	default:
		return t
	}
}

// conditionsToStrings converts map[string]any to map[string]string for metadata storage.
func conditionsToStrings(m map[string]any) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		switch v.(type) {
		case string:
			out[k] = v.(string)
		default:
			b, err := json.Marshal(v)
			if err != nil {
				out[k] = fmt.Sprintf("%v", v)
			} else {
				out[k] = string(b)
			}
		}
	}
	return out
}

// stringsToConditions converts map[string]string from metadata back to map[string]any.
func stringsToConditions(m map[string]string) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		if len(v) > 0 && (v[0] == '[' || v[0] == '{') {
			var parsed any
			if err := json.Unmarshal([]byte(v), &parsed); err == nil {
				out[k] = parsed
				continue
			}
		}
		out[k] = v
	}
	return out
}
