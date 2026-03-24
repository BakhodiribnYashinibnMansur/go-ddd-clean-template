package announcement

import (
	"testing"

	"gct/test/integration/common/setup"
)

func TestMain(m *testing.M) {
	setup.SetupTestEnvironment(m)
}

func cleanDB(t *testing.T) {
	t.Helper()
	setup.CleanDB(t)
	// Also clean announcements table
	ctx := t.Context()
	_, err := setup.TestPG.Pool.Exec(ctx, `DELETE FROM announcements`)
	if err != nil {
		t.Fatalf("cleanDB announcements error: %s", err)
	}
}
