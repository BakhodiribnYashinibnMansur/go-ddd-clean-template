package domain

type DashboardStats struct {
	TotalUsers        int64 `json:"total_users"`
	ActiveSessions    int64 `json:"active_sessions"`
	AuditLogsToday    int64 `json:"audit_logs_today"`
	SystemErrorsCount int64 `json:"system_errors_count"`
	TotalFeatureFlags int64 `json:"total_feature_flags"`
	TotalWebhooks     int64 `json:"total_webhooks"`
	TotalJobs         int64 `json:"total_jobs"`
}
