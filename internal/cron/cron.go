package cron

import (
	"context"
	"sync"
	"time"

	"gct/internal/cron/session"
	"gct/internal/shared/infrastructure/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
)

type Jobs struct {
	cron        *cron.Cron
	logger      logger.Log
	runningJobs sync.Map // jobName -> bool (running status)
	ctx         context.Context
	cancel      context.CancelFunc

	// Sub-packages
	session *session.CronJobs
}

func NewCronJobs(pool *pgxpool.Pool, redis *redis.Client, logger logger.Log) *Jobs {
	ctx, cancel := context.WithCancel(context.Background())
	return &Jobs{
		logger:  logger,
		session: session.NewSessionCronJobs(pool, redis, logger),
		ctx:     ctx,
		cancel:  cancel,
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
		c.cancel()
		ctx := c.cron.Stop()
		<-ctx.Done()
		c.logger.Infow("Cron jobs stopped successfully")
	}
}
