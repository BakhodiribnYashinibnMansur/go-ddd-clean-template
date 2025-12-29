package session

import "gct/internal/cron/model"

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
func GetSyncSessionActivityConfig() model.CronConfig {
	return model.CronConfig{
		Expression: SyncSessionActivityCronExpression,
		Name:       SyncSessionActivityCronName,
	}
}

// GetExpireOldSessionsConfig returns cron config for expiring old sessions
func GetExpireOldSessionsConfig() model.CronConfig {
	return model.CronConfig{
		Expression: ExpireOldSessionsCronExpression,
		Name:       ExpireOldSessionsCronName,
	}
}
