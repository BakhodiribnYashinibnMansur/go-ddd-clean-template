package consts

// Database table names used in SQL queries and as cache key prefixes for invalidation.
const (
	TableUsers           = "users"
	TableRole            = "role"
	TablePermission      = "permission"
	TablePolicy          = "policy"
	TableSession         = "session"
	TableRelation        = "relation"
	TableScope           = "scope"
	TableSiteSetting     = "site_settings"
	TableEndpointHistory = "endpoint_history"
	TableSystemError     = "system_errors"
	TableFunctionMetric  = "function_metrics"
	TableAuditLog        = "audit_log"
	TableIntegrations    = "integrations"
	TableAPIKeys         = "api_keys"
	TableTranslations    = "translations"
	TableDataExports     = "data_exports"
	TableFeatureFlags          = "feature_flags"
	TableFeatureFlagRuleGroups = "feature_flag_rule_groups"
	TableFeatureFlagConditions = "feature_flag_conditions"
	TableRateLimits      = "rate_limits"
	TableIPRules         = "ip_rules"


	TableAnnouncements   = "announcements"
	TableNotifications   = "notifications"
	TableFileMetadata    = "file_metadata"
	TableUserSettings    = "user_settings"
	TableEntityMetadata  = "entity_metadata"
)
