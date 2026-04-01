package integrationmod

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
	ctx := t.Context()
	_, err := setup.TestPG.Pool.Exec(ctx, `DELETE FROM api_keys; DELETE FROM integrations`)
	if err != nil {
		t.Fatalf("cleanDB integrations error: %s", err)
	}
}
