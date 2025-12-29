package cron

import (
	"time"

	"go.uber.org/zap"
)

// AddCronJobWithName adds a cron job with job wrapper for tracking execution
func (c *CronJobs) AddCronJobWithName(expression, name string, fn func()) {
	_, err := c.cron.AddFunc(expression, func() {
		c.runJobWithWrapper(name, fn)
	})
	if err != nil {
		c.logger.Error("Failed to add cron job",
			zap.String("name", name),
			zap.String("expression", expression),
			zap.Error(err),
		)
		return
	}

	c.logger.Info("Cron job registered",
		zap.String("name", name),
		zap.String("expression", expression),
	)
}

// runJobWithWrapper wraps job execution with logging and prevents duplicate runs
func (c *CronJobs) runJobWithWrapper(name string, fn func()) {
	// Check if job is already running
	if _, loaded := c.runningJobs.LoadOrStore(name, true); loaded {
		c.logger.Warn("Cronjob already running, skipping",
			zap.String("job_name", name),
		)
		return
	}

	defer c.runningJobs.Delete(name)

	startTime := time.Now()
	c.logger.Info("Cronjob started",
		zap.String("job_name", name),
		zap.Time("start_time", startTime),
	)

	// Execute the job
	fn()

	duration := time.Since(startTime)
	c.logger.Info("Cronjob completed",
		zap.String("job_name", name),
		zap.Duration("duration", duration),
	)
}
