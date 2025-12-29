package cron

import (
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"

	"gct/internal/cron/model"
	"gct/internal/cron/session"
	"gct/pkg/logger"
)

type CronJobs struct {
	cron        *cron.Cron
	logger      logger.Log
	runningJobs sync.Map // jobName -> bool (running status)
	mu          sync.Mutex

	// Sub-packages
	session *session.SessionCronJobs
}

func NewCronJobs(pool *pgxpool.Pool, redis *redis.Client, logger logger.Log) *CronJobs {
	return &CronJobs{
		logger:  logger,
		session: session.NewSessionCronJobs(pool, redis, logger),
	}
}

// addCronJob helper method to add cron job using config
func (c *CronJobs) addCronJob(cfg func() model.CronConfig, fn func()) {
	config := cfg()
	c.AddCronJobWithName(config.Expression, config.Name, fn)
}

// Start initializes and starts all cron jobs
func (c *CronJobs) Start() {
	location, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		c.logger.Errorw("Failed to load timezone, using UTC", "error", err)
		location = time.UTC
	}

	c.cron = cron.New(cron.WithLocation(location))

	// Register cron jobs
	c.addCronJob(session.GetSyncSessionActivityConfig, c.session.SyncSessionActivityToPostgres)
	c.addCronJob(session.GetExpireOldSessionsConfig, c.session.ExpireOldSessions)

	c.cron.Start()
	c.logger.Infow("Cron jobs started successfully")
}

// Stop gracefully stops all cron jobs
func (c *CronJobs) Stop() {
	if c.cron != nil {
		ctx := c.cron.Stop()
		<-ctx.Done()
		c.logger.Infow("Cron jobs stopped successfully")
	}
}
