package sessionmod

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"gct/internal/context/iam/generic/user/application/command"
	jwtpkg "gct/internal/kernel/infrastructure/security/jwt"
	"gct/test/integration/common/setup"
)

func TestMain(m *testing.M) {
	setup.SetupTestEnvironment(m)
}

func cleanDB(t *testing.T) {
	t.Helper()
	setup.CleanDB(t)
}

// testResolver is a minimal IntegrationResolver for integration tests.
type testResolver struct {
	key *rsa.PrivateKey
}

func (r *testResolver) Resolve(_ context.Context, _ string) (*command.JWTResolved, error) {
	return &command.JWTResolved{
		Name:        "gct-test",
		PrivateKey:  r.key,
		KeyID:       "test-kid",
		AccessTTL:   15 * time.Minute,
		RefreshTTL:  7 * 24 * time.Hour,
		MaxSessions: 3,
	}, nil
}

func newTestJWTConfig(t *testing.T) command.JWTConfig {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	pepper := make([]byte, 32)
	if _, err := rand.Read(pepper); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	hasher, err := jwtpkg.NewRefreshHasher(pepper)
	if err != nil {
		t.Fatalf("NewRefreshHasher: %v", err)
	}
	return command.JWTConfig{
		Issuer:        "gct-test",
		Resolver:      &testResolver{key: key},
		RefreshHasher: hasher,
	}
}
