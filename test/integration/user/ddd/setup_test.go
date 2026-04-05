package ddd

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"gct/config"
	tc "gct/test/testcontainers"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
)

var (
	testPool      *pgxpool.Pool
	testContainer testcontainers.Container
)

func TestMain(m *testing.M) {
	config.ResetTestConfig()
	cfg, err := config.NewTestConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	_, b, _, _ := runtime.Caller(0)
	rootPath := filepath.Join(filepath.Dir(b), "../../../..")

	pool, container, err := tc.RunPostgresTestContainer(cfg.Database, filepath.Join(rootPath, "migration/postgres"))
	if err != nil {
		log.Fatalf("Postgres container error: %s", err)
	}

	testPool = pool
	testContainer = container

	exitCode := m.Run()

	pool.Close()
	time.Sleep(500 * time.Millisecond)
	if err := container.Terminate(context.Background()); err != nil {
		log.Printf("failed to terminate postgres: %s", err)
	}

	os.Exit(exitCode)
}

func cleanUserTables(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	_, err := testPool.Exec(ctx, `DELETE FROM session; DELETE FROM users;`)
	if err != nil {
		t.Fatalf("cleanUserTables: %v", err)
	}
}
