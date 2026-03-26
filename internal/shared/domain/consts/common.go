package consts

// Sort order values used by query parameter binding. These are lowercase variants;
// SQL queries should use the uppercase SQLOrderAsc/SQLOrderDesc from repo.go.
const (
	OrderAsc  = "asc"
	OrderDesc = "desc"
)

const (
	FormatDate = "2006-01-02"
)

// Well-known role slugs. These are matched against the role title stored in the database,
// not the role UUID. Used primarily by middleware for fast admin-check shortcuts.
const (
	RoleUser       = "user"
	RoleAdmin      = "admin"
	RoleSuperAdmin = "super_admin"
)

const (
	AuthBearer = "Bearer "
)

const (
	DurationAuditSave = 5 // seconds
)

const (
	ServiceNameAPI = "api"
)
