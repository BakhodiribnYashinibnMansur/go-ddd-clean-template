package featureflag

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

	// Delete in FK order: conditions -> rule_groups -> flags
	for _, table := range []string{
		"feature_flag_conditions",
		"feature_flag_rule_groups",
		"feature_flags",
	} {
		_, err := setup.TestPG.Pool.Exec(ctx, `DELETE FROM `+table)
		if err != nil {
			t.Fatalf("cleanDB %s error: %s", table, err)
		}
	}
}
