package domain

import "errors"

// Domain errors for the authz bounded context.
// These sentinels are returned by aggregate methods and repository lookups.
// Application-layer handlers should match on these to produce appropriate HTTP responses.
var (
	// ErrRoleNotFound signals that no role exists for the given identifier.
	ErrRoleNotFound = errors.New("role not found")

	// ErrPermissionNotFound is returned when a permission lookup or removal targets a non-existent ID.
	// This is also returned by Role.RemovePermission when the permission is not part of the role.
	ErrPermissionNotFound = errors.New("permission not found")

	// ErrPolicyNotFound signals that no ABAC policy exists for the given identifier.
	ErrPolicyNotFound = errors.New("policy not found")

	// ErrScopeNotFound is returned by Permission.RemoveScope when the path+method pair does not exist.
	ErrScopeNotFound = errors.New("scope not found")

	// ErrDuplicatePermission prevents assigning the same permission name twice within a role.
	ErrDuplicatePermission = errors.New("duplicate permission")
)
