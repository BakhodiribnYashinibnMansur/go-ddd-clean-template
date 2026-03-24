package dashboard_test

import (
	"errors"
	"testing"

	"gct/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet_Success(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	expected := domain.DashboardStats{
		TotalUsers:        150,
		ActiveSessions:    42,
		AuditLogsToday:    87,
		SystemErrorsCount: 3,
		TotalFeatureFlags: 12,
		TotalWebhooks:     5,
		TotalJobs:         8,
	}

	repo.On("Get", ctx).Return(expected, nil)

	result, err := uc.Get(ctx)

	require.NoError(t, err)
	assert.Equal(t, expected.TotalUsers, result.TotalUsers)
	assert.Equal(t, expected.ActiveSessions, result.ActiveSessions)
	assert.Equal(t, expected.AuditLogsToday, result.AuditLogsToday)
	assert.Equal(t, expected.SystemErrorsCount, result.SystemErrorsCount)
	assert.Equal(t, expected.TotalFeatureFlags, result.TotalFeatureFlags)
	assert.Equal(t, expected.TotalWebhooks, result.TotalWebhooks)
	assert.Equal(t, expected.TotalJobs, result.TotalJobs)
	repo.AssertExpectations(t)
}

func TestGet_ZeroStats(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	expected := domain.DashboardStats{}

	repo.On("Get", ctx).Return(expected, nil)

	result, err := uc.Get(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(0), result.TotalUsers)
	assert.Equal(t, int64(0), result.ActiveSessions)
	assert.Equal(t, int64(0), result.AuditLogsToday)
	assert.Equal(t, int64(0), result.SystemErrorsCount)
	assert.Equal(t, int64(0), result.TotalFeatureFlags)
	assert.Equal(t, int64(0), result.TotalWebhooks)
	assert.Equal(t, int64(0), result.TotalJobs)
	repo.AssertExpectations(t)
}

func TestGet_RepoError(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	repo.On("Get", ctx).Return(domain.DashboardStats{}, errors.New("database unavailable"))

	result, err := uc.Get(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database unavailable")
	assert.Equal(t, domain.DashboardStats{}, result)
	repo.AssertExpectations(t)
}
