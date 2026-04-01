package usersetting

import (
	"testing"

	"gct/test/integration/common/setup"
)

func TestMain(m *testing.M) {
	setup.SetupTestEnvironment(m)
}

func cleanDB(t *testing.T) {
	t.Helper()
	ctx := t.Context()
	_, err := setup.TestPG.Pool.Exec(ctx, `DELETE FROM user_settings`)
	if err != nil {
		t.Fatalf("cleanDB user_settings error: %s", err)
	}
}
