package repository

import (
	"context"

	"gct/internal/context/iam/generic/authz/domain/entity"
	shared "gct/internal/kernel/domain"
)

// RoleRepository is the write-side repository for the Role aggregate.
// Implementations must return ErrRoleNotFound from FindByID when no row matches.
// Save persists a new role; Update persists changes to an existing one including its child permissions.
type RoleRepository interface {
	Save(ctx context.Context, role *entity.Role) error
	FindByID(ctx context.Context, id entity.RoleID) (*entity.Role, error)
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id entity.RoleID) error
	List(ctx context.Context, pagination shared.Pagination) ([]*entity.Role, int64, error)
}

// PermissionRepository is the write-side repository for Permission entities.
// Permissions may exist independently of roles (referenced via join tables), so this
// repository manages their full lifecycle separate from the Role aggregate.
type PermissionRepository interface {
	Save(ctx context.Context, perm *entity.Permission) error
	FindByID(ctx context.Context, id entity.PermissionID) (*entity.Permission, error)
	Update(ctx context.Context, perm *entity.Permission) error
	Delete(ctx context.Context, id entity.PermissionID) error
	List(ctx context.Context, pagination shared.Pagination) ([]*entity.Permission, int64, error)
}

// PolicyRepository is the write-side repository for ABAC Policy entities.
// FindByPermissionID returns all policies bound to a given permission, enabling bulk evaluation.
type PolicyRepository interface {
	Save(ctx context.Context, policy *entity.Policy) error
	FindByID(ctx context.Context, id entity.PolicyID) (*entity.Policy, error)
	Update(ctx context.Context, policy *entity.Policy) error
	Delete(ctx context.Context, id entity.PolicyID) error
	List(ctx context.Context, pagination shared.Pagination) ([]*entity.Policy, int64, error)
	FindByPermissionID(ctx context.Context, permissionID entity.PermissionID) ([]*entity.Policy, error)
}

// ScopeRepository persists Scope value objects as a global registry of API endpoints.
// Scopes are identified by their composite key (path + method), not by a UUID.
type ScopeRepository interface {
	Save(ctx context.Context, scope entity.Scope) error
	Delete(ctx context.Context, path, method string) error
	List(ctx context.Context, pagination shared.Pagination) ([]entity.Scope, int64, error)
}

// RolePermissionRepository manages the many-to-many join between roles and permissions.
// Assign and Revoke are idempotent — assigning an already-assigned pair is a no-op.
type RolePermissionRepository interface {
	Assign(ctx context.Context, roleID entity.RoleID, permissionID entity.PermissionID) error
	Revoke(ctx context.Context, roleID entity.RoleID, permissionID entity.PermissionID) error
}

// PermissionScopeRepository manages the many-to-many join between permissions and scopes.
// The composite key is (permissionID, path, method).
type PermissionScopeRepository interface {
	Assign(ctx context.Context, permissionID entity.PermissionID, path, method string) error
	Revoke(ctx context.Context, permissionID entity.PermissionID, path, method string) error
}
