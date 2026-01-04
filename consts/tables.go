package consts

const (
	// Table names matches the database table names.
	// These constants should be used as prefixes for cache keys to ensure proper invalidation.
	TableUsers      = "users"
	TableRole       = "role"
	TablePermission = "permission"
	TablePolicy     = "policy"
	TableSession    = "session"
	TableRelation   = "relation"
)
