package minio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepo_HealthCheck(t *testing.T) {
	err := testRepo.HealthCheck(testCtx)
	assert.NoError(t, err)

	// Test case where bucket doesn't exist?
	// We configured the bucket in setup, so it exists.

	// Create a repo with bad config/client to test fail case
	// The HealthCheck verifies bucket existence.
	// We can try checking a non-existent bucket using a new repo instance but same client.

	newCfg := *testRepo.config // shallow copy of struct value
	newCfg.Bucket = "random-bucket-does-not-exist"

	badRepo := New(testRepo.client, &newCfg)

	err = badRepo.HealthCheck(t.Context())
	assert.Error(t, err)
}
