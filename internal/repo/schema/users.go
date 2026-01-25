package schema

// Table name
const TableUsers = "users"

// Users table columns
const (
	UsersID           = "id"
	UsersRoleID       = "role_id"
	UsersUsername     = "username"
	UsersEmail        = "email"
	UsersPhone        = "phone"
	UsersPasswordHash = "password_hash"
	UsersSalt         = "salt"
	UsersAttributes   = "attributes"
	UsersActive       = "active"
	UsersIsApproved   = "is_approved"
	UsersLastSeen     = "last_seen"
	UsersDeletedAt    = "deleted_at"
	UsersCreatedAt    = "created_at"
	UsersUpdatedAt    = "updated_at"
)
