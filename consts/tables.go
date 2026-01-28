package consts

import "gct/internal/repo/schema"

const (
	// Table names - imported from schema package for use in cache keys
	// These constants should be used as prefixes for cache keys to ensure proper invalidation.
	TableUsers           = schema.TableUsers
	TableRole            = schema.TableRole
	TablePermission      = schema.TablePermission
	TablePolicy          = schema.TablePolicy
	TableSession         = schema.TableSession
	TableRelation        = schema.TableRelation
	TableScope           = schema.TableScope
	TableSiteSetting     = schema.TableSiteSetting
	TableEndpointHistory = schema.TableEndpointHistory
	TableSystemError     = schema.TableSystemError
	TableFunctionMetric  = schema.TableFunctionMetric
	TableAuditLog        = schema.TableAuditLog
)
