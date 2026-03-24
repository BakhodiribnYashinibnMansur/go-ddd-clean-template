package domain

import "errors"

var (
	ErrRoleNotFound        = errors.New("role not found")
	ErrPermissionNotFound  = errors.New("permission not found")
	ErrPolicyNotFound      = errors.New("policy not found")
	ErrScopeNotFound       = errors.New("scope not found")
	ErrDuplicatePermission = errors.New("duplicate permission")
)
