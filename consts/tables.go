package consts


const (
	// Table names used as prefixes for cache keys to ensure proper invalidation.
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
	TableFeatureFlags    = "feature_flags"
	TableRateLimits      = "rate_limits"
	TableIPRules         = "ip_rules"
	TableWebhooks        = "webhooks"
	TableJobs            = "jobs"
	TableAnnouncements   = "announcements"
	TableNotifications   = "notifications"
	TableEmailTemplates  = "email_templates"
	TableEmailLogs       = "email_logs"
	TableFileMetadata    = "file_metadata"
	TableUserSettings    = "user_settings"
)
