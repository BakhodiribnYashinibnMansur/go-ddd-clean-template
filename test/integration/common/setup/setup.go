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
	tc "gct/internal/shared/infrastructure/container"
	dbPostgres "gct/internal/shared/infrastructure/db/postgres"
	"gct/internal/shared/infrastructure/logger"
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
	// Base is .../gct/test/integration/common/setup/setup.go
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

	pgPool, pgC, err := tc.RunPostgresTestContainer(cfg.Database, filepath.Join(rootPath, "migrations/postgres"))
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

	var err error
	for range 5 {
		_, err = TestPG.Pool.Exec(ctx, `
			DELETE FROM file_metadata;
			DELETE FROM endpoint_history;
			DELETE FROM audit_log;
			DELETE FROM user_relation;
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

func setupKeys(cfg *config.Config) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("rsa.GenerateKey error: %s", err)
	}
	privBytes := x509.MarshalPKCS1PrivateKey(key)
	privPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
	cfg.JWT.PrivateKey = string(privPem)
	pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		log.Fatalf("x509.MarshalPKIXPublicKey error: %s", err)
	}
	pubPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: pubBytes})
	cfg.JWT.PublicKey = string(pubPem)
	cfg.JWT.Issuer = "gct-integration"
}
