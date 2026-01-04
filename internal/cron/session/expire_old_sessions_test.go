package session

import (
	"errors"
	"testing"

	"gct/pkg/logger"
	"github.com/go-redis/redismock/v9"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCronJobs_ExpireOldSessions(t *testing.T) {
	t.Parallel()

	// Setup mocks
	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	redisClient, _ := redismock.NewClientMock()
	log := logger.New("debug")

	c := &CronJobs{
		pool:   mockPool,
		redis:  redisClient,
		logger: log,
	}

	// Test case: Success
	mockPool.ExpectExec(`UPDATE sessions`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 10))

	c.ExpireOldSessions()

	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestCronJobs_ExpireOldSessions_Error(t *testing.T) {
	t.Parallel()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	redisClient, _ := redismock.NewClientMock()
	log := logger.New("debug")

	c := &CronJobs{
		pool:   mockPool,
		redis:  redisClient,
		logger: log,
	}

	// Test case: Error
	mockPool.ExpectExec(`UPDATE sessions`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnError(errors.New("db error"))

	c.ExpireOldSessions()

	assert.NoError(t, mockPool.ExpectationsWereMet())
}
