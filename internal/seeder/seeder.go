// Package seeder provides data seeding functionality for test and development environments.
package seeder

import (
	"context"
	"fmt"
	"time"

	"gct/config"
	"gct/internal/repo"
	"gct/pkg/logger"
	"github.com/brianvoe/gofakeit/v7"
	"go.uber.org/zap"
)

// Seeder manages data seeding operations.
type Seeder struct {
	repo   *repo.Repo
	logger logger.Log
	cfg    *config.Config
}

// New creates a new Seeder instance.
func New(repo *repo.Repo, logger logger.Log, cfg *config.Config) *Seeder {
	return &Seeder{
		repo:   repo,
		logger: logger,
		cfg:    cfg,
	}
}

// Seed executes all seeding operations.
func (s *Seeder) Seed(ctx context.Context, customCounts map[string]int) error {
	if !s.cfg.Seeder.IsEnabled() && customCounts == nil {
		s.logger.Infoc(ctx, "Seeder is disabled, skipping...")
		return nil
	}

	s.logger.Infoc(ctx, "Starting data seeding...")
	startTime := time.Now()

	// Set seed for reproducible data
	seed := s.cfg.Seeder.Seed
	// Check if seed is provided in map via special key
	if customCounts != nil {
		if val, ok := customCounts["seed"]; ok {
			seed = int64(val)
		}
	}

	if seed != 0 {
		gofakeit.Seed(seed)
		s.logger.Infoc(ctx, "Using custom seed for reproducible data", zap.Int64("seed", seed))
	} else {
		gofakeit.Seed(0) // Random seed
		s.logger.Infoc(ctx, "Using random seed")
	}

	// Clear existing data if requested
	// Check overrides first
	shouldClear := s.cfg.Seeder.ShouldClearData()
	if customCounts != nil {
		if val, ok := customCounts["clear_data"]; ok {
			shouldClear = val == 1
		}
	}

	if shouldClear {
		if err := s.clearData(ctx); err != nil {
			return fmt.Errorf("failed to clear data: %w", err)
		}
	}

	// Resolve counts
	permCount := s.getCount(customCounts, "permissions", s.cfg.Seeder.PermissionsCount)
	roleCount := s.getCount(customCounts, "roles", s.cfg.Seeder.RolesCount)
	userCount := s.getCount(customCounts, "users", s.cfg.Seeder.UsersCount)
	policyCount := s.getCount(customCounts, "policies", s.cfg.Seeder.PoliciesCount)

	// Seed data in order (respecting foreign key constraints)
	if err := s.seedPermissions(ctx, permCount); err != nil {
		return fmt.Errorf("failed to seed permissions: %w", err)
	}

	if err := s.seedRoles(ctx, roleCount); err != nil {
		return fmt.Errorf("failed to seed roles: %w", err)
	}

	if err := s.seedUsers(ctx, userCount); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	if err := s.seedRolePermissions(ctx); err != nil {
		return fmt.Errorf("failed to seed role permissions: %w", err)
	}

	if err := s.seedPolicies(ctx, policyCount); err != nil {
		return fmt.Errorf("failed to seed policies: %w", err)
	}

	duration := time.Since(startTime)
	s.logger.Infoc(ctx, "Data seeding completed successfully",
		zap.Duration("duration", duration),
	)

	return nil
}

// clearData removes all existing data from tables.
func (s *Seeder) clearData(ctx context.Context) error {
	s.logger.Warnc(ctx, "Clearing existing data...")

	// Order matters: delete in reverse order of foreign key dependencies
	tables := []string{
		"user_relation",
		"role_permission",
		"session",
		"audit_log",
		"policy",
		"users",
		"role",
		"permission",
	}

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		if _, err := s.repo.Persistent.Postgres.DB.Pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
		s.logger.Infoc(ctx, "Table truncated", zap.String("table", table))
	}

	return nil
}

func (s *Seeder) getCount(counts map[string]int, key string, defaultVal int) int {
	if counts != nil {
		if val, ok := counts[key]; ok {
			return val
		}
	}
	return defaultVal
}
