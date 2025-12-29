package model

// CronConfig represents configuration for a cron job
type CronConfig struct {
	Expression string // Cron expression (e.g., "*/5 * * * *")
	Name       string // Job name for logging
}
