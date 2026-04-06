package sessionmod

import (
	"testing"

	"gct/internal/context/iam/generic/user/application/command"
	"gct/test/integration/common/setup"
)

func TestMain(m *testing.M) {
	setup.SetupTestEnvironment(m)
}

func cleanDB(t *testing.T) {
	t.Helper()
	setup.CleanDB(t)
}

func newTestJWTConfig(t *testing.T) command.JWTConfig {
	t.Helper()
	return command.JWTConfig{
		Issuer: "gct-test",
	}
}
