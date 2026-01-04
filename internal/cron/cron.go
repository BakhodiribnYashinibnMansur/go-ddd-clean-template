package cron

import (
	"sync"
	"time"

	"gct/internal/cron/session"
	"gct/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
)

type Jobs struct {
	cron        *cron.Cron
	logger      logger.Log
	runningJobs sync.Map // jobName -> bool (running status)

	// Sub-packages
	session *session.CronJobs
}

func NewCronJobs(pool *pgxpool.Pool, redis *redis.Client, logger logger.Log) *Jobs {
	return &Jobs{
		logger:  logger,
		session: session.NewSessionCronJobs(pool, redis, logger),
	}
}

// Start initializes and starts all cron jobs
func (c *Jobs) Start() {
	location, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		c.logger.Errorw("Failed to load timezone, using UTC", "error", err)
		location = time.UTC
	}

	c.cron = cron.New(cron.WithLocation(location))

	// Register cron jobs
	syncCfg := session.GetSyncSessionActivityConfig()
	c.AddCronJobWithName(syncCfg.Expression, syncCfg.Name, c.session.SyncSessionActivityToPostgres)

	expireCfg := session.GetExpireOldSessionsConfig()
	c.AddCronJobWithName(expireCfg.Expression, expireCfg.Name, c.session.ExpireOldSessions)

	c.cron.Start()
	c.logger.Infow("Cron jobs started successfully")
}

// Stop gracefully stops all cron jobs
func (c *Jobs) Stop() {
	if c.cron != nil {
		ctx := c.cron.Stop()
		<-ctx.Done()
		c.logger.Infow("Cron jobs stopped successfully")
	}
}
