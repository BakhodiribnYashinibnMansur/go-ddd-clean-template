package domain

import shared "gct/internal/kernel/domain"

// Domain errors for the authz bounded context.
// These sentinels are returned by aggregate methods and repository lookups.
// Application-layer handlers should match on these to produce appropriate HTTP responses.
var (
	// ErrRoleNotFound signals that no role exists for the given identifier.
	ErrRoleNotFound = shared.NewDomainError("AUTHZ_ROLE_NOT_FOUND", "role not found")

	// ErrPermissionNotFound is returned when a permission lookup or removal targets a non-existent ID.
	// This is also returned by Role.RemovePermission when the permission is not part of the role.
	ErrPermissionNotFound = shared.NewDomainError("AUTHZ_PERMISSION_NOT_FOUND", "permission not found")

	// ErrPolicyNotFound signals that no ABAC policy exists for the given identifier.
	ErrPolicyNotFound = shared.NewDomainError("AUTHZ_POLICY_NOT_FOUND", "policy not found")

	// ErrScopeNotFound is returned by Permission.RemoveScope when the path+method pair does not exist.
	ErrScopeNotFound = shared.NewDomainError("AUTHZ_SCOPE_NOT_FOUND", "scope not found")

	// ErrDuplicatePermission prevents assigning the same permission name twice within a role.
	ErrDuplicatePermission = shared.NewDomainError("AUTHZ_DUPLICATE_PERMISSION", "duplicate permission")
)
