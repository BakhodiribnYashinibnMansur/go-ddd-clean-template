package application

// OverviewView aggregates the top-level counts shown on the admin overview page.
type OverviewView struct {
	TotalUsers        int64 `json:"total_users"`
	ActiveSessions    int64 `json:"active_sessions"`
	AuditLogsToday    int64 `json:"audit_logs_today"`
	SystemErrorsCount int64 `json:"system_errors_count"`
	TotalFeatureFlags int64 `json:"total_feature_flags"`
}

// UserStatsView breaks down users by lifecycle and role.
type UserStatsView struct {
	Total   int64            `json:"total"`
	Deleted int64            `json:"deleted"`
	ByRole  map[string]int64 `json:"by_role"`
}

// SessionStatsView breaks down sessions by state.
type SessionStatsView struct {
	Active  int64 `json:"active"`
	Expired int64 `json:"expired"`
	Revoked int64 `json:"revoked"`
}

// ErrorStatsView breaks down system errors by resolution state and recency.
type ErrorStatsView struct {
	Unresolved int64 `json:"unresolved"`
	Resolved   int64 `json:"resolved"`
	Last24h    int64 `json:"last_24h"`
}

// AuditStatsView breaks down audit log entries by recency.
type AuditStatsView struct {
	Today     int64 `json:"today"`
	Last7Days int64 `json:"last_7_days"`
	Total     int64 `json:"total"`
}

// SecurityStatsView aggregates counts for security-related tables.
type SecurityStatsView struct {
	IPRules    int64 `json:"ip_rules"`
	RateLimits int64 `json:"rate_limits"`
}

// FeatureFlagStatsView breaks down feature flags by active state.
type FeatureFlagStatsView struct {
	Total    int64 `json:"total"`
	Enabled  int64 `json:"enabled"`
	Disabled int64 `json:"disabled"`
}

// ContentStatsView aggregates counts for content tables.
type ContentStatsView struct {
	Announcements int64 `json:"announcements"`
	Notifications int64 `json:"notifications"`
	FileMetadata  int64 `json:"file_metadata"`
	Translations  int64 `json:"translations"`
}

// IntegrationStatsView aggregates counts for integration tables.
type IntegrationStatsView struct {
	Integrations int64 `json:"integrations"`
	APIKeys      int64 `json:"api_keys"`
}
