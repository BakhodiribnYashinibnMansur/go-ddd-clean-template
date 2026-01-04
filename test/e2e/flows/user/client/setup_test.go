package client

import (
	"testing"

	"gct/test/e2e/common/setup"
)

// Use shared setup from parent package
// Vars are accessed directly from setup package

func TestMain(m *testing.M) {
	// Delegate to shared setup
	setup.SetupTestEnvironment(m)
}

func cleanDB(t *testing.T) {
	setup.CleanDB(t)
}
