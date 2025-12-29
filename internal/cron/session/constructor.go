package session

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"gct/pkg/logger"
)

type SessionCronJobs struct {
	pool   *pgxpool.Pool
	redis  *redis.Client
	logger logger.Log
}

func NewSessionCronJobs(pool *pgxpool.Pool, redis *redis.Client, logger logger.Log) *SessionCronJobs {
	return &SessionCronJobs{
		pool:   pool,
		redis:  redis,
		logger: logger,
	}
}
