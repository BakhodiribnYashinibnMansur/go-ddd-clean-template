package session

import (
	"errors"
	"testing"
	"time"

	"gct/pkg/logger"

	"github.com/go-redis/redismock/v9"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sessionPattern = "session_last_activity:*"

func TestCronJobs_SyncSessionActivityToPostgres(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	redisClient, redisMock := redismock.NewClientMock()
	log := logger.New("debug")

	c := &CronJobs{
		pool:   mockPool,
		redis:  redisClient,
		logger: log,
	}

	sessionID := "550e8400-e29b-41d4-a716-446655440000"
	key := "session_last_activity:" + sessionID
	lastActivity := time.Now().Truncate(time.Second)

	// Redis expectations
	redisMock.ExpectScan(0, sessionPattern, 0).SetVal([]string{key}, 0)
	redisMock.ExpectGet(key).SetVal(lastActivity.Format(time.RFC3339))

	// Postgres expectations
	mockPool.ExpectBegin()
	mockPool.ExpectExec(`CREATE TEMP TABLE temp_session_activity`).
		WillReturnResult(pgxmock.NewResult("CREATE", 0))
	mockPool.ExpectCopyFrom(
		pgx.Identifier{"temp_session_activity"},
		[]string{"session_id", "last_activity"},
	).WillReturnResult(1)
	mockPool.ExpectExec(`UPDATE session AS s`).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mockPool.ExpectCommit()

	c.SyncSessionActivityToPostgres()

	assert.NoError(t, redisMock.ExpectationsWereMet())
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestCronJobs_SyncSessionActivityToPostgres_NoData(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	redisClient, redisMock := redismock.NewClientMock()
	log := logger.New("debug")

	c := &CronJobs{
		pool:   mockPool,
		redis:  redisClient,
		logger: log,
	}

	// Redis expectations: No keys found
	redisMock.ExpectScan(0, sessionPattern, 0).SetVal([]string{}, 0)

	c.SyncSessionActivityToPostgres()

	assert.NoError(t, redisMock.ExpectationsWereMet())
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestCronJobs_SyncSessionActivityToPostgres_RedisError(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	redisClient, redisMock := redismock.NewClientMock()
	log := logger.New("debug")

	c := &CronJobs{
		pool:   mockPool,
		redis:  redisClient,
		logger: log,
	}

	// Redis expectations: Scan error
	redisMock.ExpectScan(0, sessionPattern, 0).SetErr(errors.New("redis scan error"))

	c.SyncSessionActivityToPostgres()

	assert.NoError(t, redisMock.ExpectationsWereMet())
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
