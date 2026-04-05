package sessionmod

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"gct/internal/context/iam/user/application/command"
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
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	return command.JWTConfig{
		PrivateKey: key,
		Issuer:     "gct-test",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}
}
