package session

import (
	"context"

	"gct/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
)

type PgxPool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	Close()
}

type CronJobs struct {
	pool   PgxPool
	redis  *redis.Client
	logger logger.Log
}

func NewSessionCronJobs(pool PgxPool, redis *redis.Client, logger logger.Log) *CronJobs {
	return &CronJobs{
		pool:   pool,
		redis:  redis,
		logger: logger,
	}
}
