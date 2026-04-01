package access

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

	for _, table := range []string{
		"permission_scope",
		"role_permission",
		"policy",
		"scope",
		"permission",
		"role",
	} {
		_, err := setup.TestPG.Pool.Exec(ctx, "DELETE FROM "+table)
		if err != nil {
			t.Fatalf("cleanDB %s: %s", table, err)
		}
	}
}
