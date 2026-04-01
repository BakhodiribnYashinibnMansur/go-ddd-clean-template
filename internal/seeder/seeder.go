// Package seeder provides data seeding functionality for test and development environments.
package seeder

import (
	"context"
	"fmt"
	"time"

	"gct/config"
	"gct/internal/shared/infrastructure/logger"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Seeder manages data seeding operations.
type Seeder struct {
	pool   *pgxpool.Pool
	logger logger.Log
	cfg    *config.Config
}

// New creates a new Seeder instance.
func New(pool *pgxpool.Pool, logger logger.Log, cfg *config.Config) *Seeder {
	return &Seeder{
		pool:   pool,
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

	// Phase 1: Independent tables
	siteSettingsCount := s.getCount(customCounts, "site_settings", s.cfg.Seeder.SiteSettingsCount)
	if err := s.seedSiteSettings(ctx, siteSettingsCount); err != nil {
		return fmt.Errorf("failed to seed site settings: %w", err)
	}

	errorCodesCount := s.getCount(customCounts, "error_codes", s.cfg.Seeder.ErrorCodesCount)
	if err := s.seedErrorCodes(ctx, errorCodesCount); err != nil {
		return fmt.Errorf("failed to seed error codes: %w", err)
	}

	featureFlagsCount := s.getCount(customCounts, "feature_flags", s.cfg.Seeder.FeatureFlagsCount)
	if err := s.seedFeatureFlags(ctx, featureFlagsCount); err != nil {
		return fmt.Errorf("failed to seed feature flags: %w", err)
	}

	rateLimitsCount := s.getCount(customCounts, "rate_limits", s.cfg.Seeder.RateLimitsCount)
	if err := s.seedRateLimits(ctx, rateLimitsCount); err != nil {
		return fmt.Errorf("failed to seed rate limits: %w", err)
	}

	ipRulesCount := s.getCount(customCounts, "ip_rules", s.cfg.Seeder.IPRulesCount)
	if err := s.seedIPRules(ctx, ipRulesCount); err != nil {
		return fmt.Errorf("failed to seed ip rules: %w", err)
	}

	functionMetricsCount := s.getCount(customCounts, "function_metrics", s.cfg.Seeder.FunctionMetricsCount)
	if err := s.seedFunctionMetrics(ctx, functionMetricsCount); err != nil {
		return fmt.Errorf("failed to seed function metrics: %w", err)
	}

	// Phase 2: Depends on users
	integrationsCount := s.getCount(customCounts, "integrations", s.cfg.Seeder.IntegrationsCount)
	if err := s.seedIntegrations(ctx, integrationsCount); err != nil {
		return fmt.Errorf("failed to seed integrations: %w", err)
	}

	fileMetadataCount := s.getCount(customCounts, "file_metadata", s.cfg.Seeder.FileMetadataCount)
	if err := s.seedFileMetadata(ctx, fileMetadataCount); err != nil {
		return fmt.Errorf("failed to seed file metadata: %w", err)
	}

	translationsCount := s.getCount(customCounts, "translations", s.cfg.Seeder.TranslationsCount)
	if err := s.seedTranslations(ctx, translationsCount); err != nil {
		return fmt.Errorf("failed to seed translations: %w", err)
	}

	announcementsCount := s.getCount(customCounts, "announcements", s.cfg.Seeder.AnnouncementsCount)
	if err := s.seedAnnouncements(ctx, announcementsCount); err != nil {
		return fmt.Errorf("failed to seed announcements: %w", err)
	}

	notificationsCount := s.getCount(customCounts, "notifications", s.cfg.Seeder.NotificationsCount)
	if err := s.seedNotifications(ctx, notificationsCount); err != nil {
		return fmt.Errorf("failed to seed notifications: %w", err)
	}

	// Phase 3: Depends on users + sessions + policies
	auditLogsCount := s.getCount(customCounts, "audit_logs", s.cfg.Seeder.AuditLogsCount)
	if err := s.seedAuditLogs(ctx, auditLogsCount); err != nil {
		return fmt.Errorf("failed to seed audit logs: %w", err)
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
		// Reverse FK order — dependent tables first
		"function_metrics",
		"endpoint_history",
		"audit_log",
		"notifications",
		"announcements",
		"translations",
		"file_metadata",
		"api_keys",
		"integrations",
		"ip_rules",
		"rate_limits",
		"feature_flags",
		"error_code",
		"site_settings",
		// Original tables
		"user_relation",
		"role_permission",
		"session",
		"policy",
		"users",
		"role",
		"permission",
	}

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)
		if _, err := s.pool.Exec(ctx, query); err != nil {
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
