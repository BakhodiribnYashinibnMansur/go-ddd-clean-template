package schema

// Table name
const TableRole = "role"

// Role table columns
const (
	RoleID        = "id"
	RoleName      = "name"
	RoleCreatedAt = "created_at"
)

// Table name
const TableRolePermission = "role_permission"

// RolePermission table columns
const (
	RolePermissionRoleID       = "role_id"
	RolePermissionPermissionID = "permission_id"
	RolePermissionCreatedAt    = "created_at"
)
