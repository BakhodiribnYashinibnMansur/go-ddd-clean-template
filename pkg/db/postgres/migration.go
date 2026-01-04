package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"gct/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// ErrMigrationDirNotFound is returned when migrations directory doesn't exist.
var ErrMigrationDirNotFound = errors.New("migrations directory not found")

// Migration - goose migration manager.
type Migration struct {
	db  *sql.DB
	dir string
}

// NewMigration creates a new goose migration manager.
func NewMigration(cfg config.Postgres, dir string) (*Migration, error) {
	db, err := sql.Open("pgx", cfg.URL())
	if err != nil {
		return nil, fmt.Errorf("postgres - NewMigration - sql.Open: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		db.Close()
		return nil, fmt.Errorf("postgres - NewMigration - goose.SetDialect: %w", err)
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		db.Close()
		return nil, fmt.Errorf("%w: %s", ErrMigrationDirNotFound, dir)
	}

	return &Migration{
		db:  db,
		dir: dir,
	}, nil
}

// Close closes the database connection.
func (m *Migration) Close() error {
	return m.db.Close()
}

// Up runs all available migrations.
func (m *Migration) Up() error {
	if err := goose.Up(m.db, m.dir); err != nil {
		return fmt.Errorf("postgres - Migration - Up: %w", err)
	}
	return nil
}

// Down rolls back a single migration from the current version.
func (m *Migration) Down() error {
	if err := goose.Down(m.db, m.dir); err != nil {
		return fmt.Errorf("postgres - Migration - Down: %w", err)
	}
	return nil
}

// Redo rolls back the most recently applied migration, then runs it again.
func (m *Migration) Redo() error {
	if err := goose.Redo(m.db, m.dir); err != nil {
		return fmt.Errorf("postgres - Migration - Redo: %w", err)
	}
	return nil
}

// Reset rolls back all migrations.
func (m *Migration) Reset() error {
	if err := goose.Reset(m.db, m.dir); err != nil {
		return fmt.Errorf("postgres - Migration - Reset: %w", err)
	}
	return nil
}

// Status prints the status of all migrations.
func (m *Migration) Status() error {
	if err := goose.Status(m.db, m.dir); err != nil {
		return fmt.Errorf("postgres - Migration - Status: %w", err)
	}
	return nil
}

// Version prints the current version of the database.
func (m *Migration) Version() error {
	if err := goose.Version(m.db, m.dir); err != nil {
		return fmt.Errorf("postgres - Migration - Version: %w", err)
	}
	return nil
}

// UpTo migrates up to a specific version.
func (m *Migration) UpTo(version int64) error {
	if err := goose.UpTo(m.db, m.dir, version); err != nil {
		return fmt.Errorf("postgres - Migration - UpTo: %w", err)
	}
	return nil
}

// DownTo rolls back migrations down to a specific version.
func (m *Migration) DownTo(version int64) error {
	if err := goose.DownTo(m.db, m.dir, version); err != nil {
		return fmt.Errorf("postgres - Migration - DownTo: %w", err)
	}
	return nil
}
