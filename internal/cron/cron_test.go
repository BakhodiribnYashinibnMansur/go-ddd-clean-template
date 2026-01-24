package cron

import (
	"context"
	"testing"

	"gct/pkg/logger"
	getCron "github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
)

func TestNewCronJobs(t *testing.T) {
	t.Parallel()
	log := logger.New("debug")
	c := NewCronJobs(nil, nil, log)
	assert.NotNil(t, c)
	assert.NotNil(t, c.session)
}

func TestJobs_Lifecycle(t *testing.T) {
	t.Parallel()
	log := logger.New("debug")
	c := NewCronJobs(nil, nil, log)
	assert.NotNil(t, c)
}

func TestJobs_AddCronJobWithName(t *testing.T) {
	t.Parallel()
	log := logger.New("debug")
	c := NewCronJobs(nil, nil, log)

	// Manually inject cron instance to test adding jobs
	c.cron = getCron.New()

	executed := false
	c.AddCronJobWithName("* * * * *", "test_job", func(ctx context.Context) {
		executed = true
	})

	entries := c.cron.Entries()
	assert.Len(t, entries, 1)

	c.runJobWithWrapper("test_job", func(ctx context.Context) {
		executed = true
	})
	assert.True(t, executed)
}
