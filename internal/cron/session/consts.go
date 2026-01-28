package session

// CronConfig matches gct/internal/cron.CronConfig but defined here to avoid import cycles.
type CronConfig struct {
	Expression string
	Name       string
}

const (
	// Cron expressions
	SyncSessionActivityCronExpression = "*/5 * * * *" // Har 5 daqiqada
	ExpireOldSessionsCronExpression   = "0 2 * * *"   // Har kuni soat 2:00 da

	// Cron names
	SyncSessionActivityCronName = "SyncSessionActivityToPostgres"
	ExpireOldSessionsCronName   = "ExpireOldSessions"

	// Session expiration settings
	SessionExpirationDays = 30 // 1 oy (30 kun)
)

// GetSyncSessionActivityConfig returns cron config for syncing session activity
func GetSyncSessionActivityConfig() CronConfig {
	return CronConfig{
		Expression: SyncSessionActivityCronExpression,
		Name:       SyncSessionActivityCronName,
	}
}

// GetExpireOldSessionsConfig returns cron config for expiring old sessions
func GetExpireOldSessionsConfig() CronConfig {
	return CronConfig{
		Expression: ExpireOldSessionsCronExpression,
		Name:       ExpireOldSessionsCronName,
	}
}
