package setup

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"gct/config"
	tc "gct/test/testcontainers"
	dbPostgres "gct/internal/kernel/infrastructure/db/postgres"
	"gct/internal/kernel/infrastructure/logger"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
)

var (
	TestPG    *dbPostgres.Postgres
	TestRedis *redis.Client
	TestMinio *minio.Client
	TestCfg   *config.Config
	rootPath  string
)

func init() {
	_, b, _, _ := runtime.Caller(0)
	// Base is .../gct/test/e2e/common/setup/setup.go
	// We want .../gct
	rootPath = filepath.Join(filepath.Dir(b), "../../../..")
}

func SetupTestEnvironment(m *testing.M) {
	ctx := context.Background()

	// 1. Use test configuration instead of loading .env file
	config.ResetTestConfig()
	cfg, err := config.NewTestConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	pgPool, pgC, err := tc.RunPostgresTestContainer(cfg.Database, filepath.Join(rootPath, "migration/postgres"))
	if err != nil {
		log.Fatalf("Postgres container error: %s", err)
	}

	// Extract config from Pool
	pgConfig := pgPool.Config().ConnConfig
	cfg.Database.Postgres.Host = pgConfig.Host
	cfg.Database.Postgres.Port = int(pgConfig.Port)
	cfg.Database.Postgres.User = pgConfig.User
	cfg.Database.Postgres.Password = pgConfig.Password
	cfg.Database.Postgres.Name = pgConfig.Database
	cfg.Database.Postgres.SSLMode = "disable"

	// 2. Start Redis Container
	rStoreCfg := config.RedisStore{
		Host:     cfg.Database.Redis.Host,
		Port:     strconv.Itoa(cfg.Database.Redis.Port),
		Password: cfg.Database.Redis.Password,
		DB:       0,
	}
	rClient, rC, err := tc.RunRedisTestContainer(rStoreCfg)
	if err != nil {
		log.Fatalf("Redis container error: %s", err)
	}

	// Extract config from Redis Client
	rOpt := rClient.Options()
	host, port, _ := net.SplitHostPort(rOpt.Addr)
	cfg.Database.Redis.Host = host
	_, _ = fmt.Sscanf(port, "%d", &cfg.Database.Redis.Port)

	TestRedis = rClient

	// 3. Start Minio Container
	mClient, mC, err := tc.RunMinioTestContainer(cfg.Minio)
	if err != nil {
		log.Fatalf("Minio container error: %s", err)
	}
	TestMinio = mClient

	setupKeys(cfg)
	TestCfg = cfg

	l := logger.New("debug")

	pg, err := dbPostgres.New(ctx, cfg.App.Environment, cfg.Database.Postgres, l)
	if err != nil {
		log.Fatalf("Postgres init error: %s", err)
	}
	TestPG = pg

	// Run tests
	exitCode := m.Run()

	// Cleanup
	time.Sleep(500 * time.Millisecond)
	pg.Close()
	rClient.Close()
	pgPool.Close()

	if err := pgC.Terminate(ctx); err != nil {
		log.Printf("failed to terminate postgres: %s", err)
	}
	if err := rC.Terminate(ctx); err != nil {
		log.Printf("failed to terminate redis: %s", err)
	}
	if err := mC.Terminate(ctx); err != nil {
		log.Printf("failed to terminate minio: %s", err)
	}

	os.Exit(exitCode)
}

func CleanDB(t *testing.T) {
	t.Helper()
	ctx := t.Context()

	// Use a retry loop because asynchronous background tasks (like EndpointHistory middleware)
	// might still be inserting rows that reference session/users while we try to clean up.
	var err error
	for range 5 {
		_, err = TestPG.Pool.Exec(ctx, `
			DELETE FROM endpoint_history;
			DELETE FROM audit_log;
			DELETE FROM user_relation;
			DELETE FROM notification;
			DELETE FROM announcement;
			DELETE FROM translation;
			DELETE FROM site_setting;
			DELETE FROM feature_flag_rule_group;
			DELETE FROM feature_flag;
			DELETE FROM data_export;
			DELETE FROM user_setting;
			DELETE FROM error_code;
			DELETE FROM rate_limit_rule;
			DELETE FROM ip_rule;
			DELETE FROM role_permission;
			DELETE FROM permission_scope;
			DELETE FROM policy;
			DELETE FROM scope;
			DELETE FROM permission;
			DELETE FROM role;
			DELETE FROM session;
			DELETE FROM users;
		`)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err != nil {
		t.Fatalf("cleanDB error after retries: %s", err)
	}

	if TestRedis != nil {
		TestRedis.FlushAll(ctx)
	}
}

// setupKeys materialises the minimum JWT config for the e2e harness. It
// generates per-integration RSA PEM files into a temp KeysDir and installs
// the shared peppers expected by the JWT + Integration packages. The keyring
// service loads these files at app boot.
func setupKeys(cfg *config.Config) {
	dir, err := os.MkdirTemp("", "gct-e2e-keys-*")
	if err != nil {
		log.Fatalf("e2e setupKeys: MkdirTemp: %s", err)
	}
	cfg.JWT.KeysDir = dir
	cfg.JWT.Issuer = "gct-e2e"
	cfg.JWT.KeyBits = 2048

	// Fixed 48-byte peppers (base64 std) — valid for config validation.
	cfg.JWT.RefreshPepper = "dGVzdC1lMmUtcmVmcmVzaC1wZXBwZXItbWluLTMyLWJ5dGVzLWxvbmctc29tZS1sbw"
	cfg.JWT.APIKeyPepper = "dGVzdC1lMmUtYXBpLWtleS1wZXBwZXItbWluLTMyLWJ5dGVzLWxvbmctc29tZS1zbw"

	// Pre-generate keys for the three seeded integrations so the keyring
	// service finds files waiting for it.
	for _, name := range []string{"gct-admin", "gct-client", "gct-mobile"} {
		k, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Fatalf("e2e setupKeys: rsa.GenerateKey: %s", err)
		}
		privDER, err := x509.MarshalPKCS8PrivateKey(k)
		if err != nil {
			log.Fatalf("e2e setupKeys: MarshalPKCS8PrivateKey: %s", err)
		}
		privPem := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})
		if err := os.WriteFile(filepath.Join(dir, name+"_private.pem"), privPem, 0o600); err != nil {
			log.Fatalf("e2e setupKeys: write private: %s", err)
		}
		pubDER, err := x509.MarshalPKIXPublicKey(&k.PublicKey)
		if err != nil {
			log.Fatalf("e2e setupKeys: MarshalPKIXPublicKey: %s", err)
		}
		pubPem := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
		if err := os.WriteFile(filepath.Join(dir, name+"_public.pem"), pubPem, 0o644); err != nil {
			log.Fatalf("e2e setupKeys: write public: %s", err)
		}
	}
}
