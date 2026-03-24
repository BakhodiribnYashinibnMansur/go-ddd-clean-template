package dashboard

import (
	"errors"
	"testing"

	"gct/internal/shared/infrastructure/logger"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Get_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	// 7 sequential COUNT queries
	mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE deleted_at = 0`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(150)))

	mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM session WHERE expires_at > NOW\(\)`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(42)))

	mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM audit_log WHERE created_at >= NOW\(\)::date`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(87)))

	mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM system_errors WHERE is_resolved = false`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(3)))

	mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM feature_flags WHERE deleted_at IS NULL AND is_active = true`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(12)))

	mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM webhooks WHERE deleted_at IS NULL AND is_active = true`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(5)))

	mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM jobs WHERE is_active = true`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(8)))

	repo := &Repo{
		pool:   mockPool,
		logger: logger.New("debug"),
	}

	stats, err := repo.Get(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(150), stats.TotalUsers)
	assert.Equal(t, int64(42), stats.ActiveSessions)
	assert.Equal(t, int64(87), stats.AuditLogsToday)
	assert.Equal(t, int64(3), stats.SystemErrorsCount)
	assert.Equal(t, int64(12), stats.TotalFeatureFlags)
	assert.Equal(t, int64(5), stats.TotalWebhooks)
	assert.Equal(t, int64(8), stats.TotalJobs)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Get_ZeroStats(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM session`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM audit_log`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM system_errors`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM feature_flags`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM webhooks`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM jobs`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))

	repo := &Repo{
		pool:   mockPool,
		logger: logger.New("debug"),
	}

	stats, err := repo.Get(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(0), stats.TotalUsers)
	assert.Equal(t, int64(0), stats.ActiveSessions)
	assert.Equal(t, int64(0), stats.AuditLogsToday)
	assert.Equal(t, int64(0), stats.SystemErrorsCount)
	assert.Equal(t, int64(0), stats.TotalFeatureFlags)
	assert.Equal(t, int64(0), stats.TotalWebhooks)
	assert.Equal(t, int64(0), stats.TotalJobs)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Get_FirstQueryError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnError(errors.New("connection refused"))

	repo := &Repo{
		pool:   mockPool,
		logger: logger.New("debug"),
	}

	stats, err := repo.Get(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Equal(t, int64(0), stats.TotalUsers)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Get_MiddleQueryError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	// First 3 queries succeed
	mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(100)))

	mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM session`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(50)))

	mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM audit_log`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(25)))

	// 4th query fails
	mockPool.ExpectQuery(`SELECT COUNT\(\*\) FROM system_errors`).
		WillReturnError(errors.New("table not found"))

	repo := &Repo{
		pool:   mockPool,
		logger: logger.New("debug"),
	}

	stats, err := repo.Get(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "table not found")
	assert.Equal(t, int64(0), stats.TotalUsers) // returns zero struct on error
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
