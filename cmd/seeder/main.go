// Package main provides a utility to trigger database seeding jobs via the background task queue.
// This is useful for populating dev/staging environments with mock data using the Asynq worker system.
package main

import (
	"context"
	"fmt"
	"os"

	"gct/config"
	"gct/internal/platform/infrastructure/asynq"
	"gct/internal/platform/infrastructure/logger"

	"go.uber.org/zap"
)

// main connects to the task queue and enqueues a database seeding instruction.
func main() {
	ctx := context.Background()

	// 1. Load application configuration.
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize logger for progress tracking.
	l := logger.New(cfg.Log.Level)

	// 3. Initialize the Asynq Client.
	// The client connects to Redis to communicate with background workers.
	client := asynq.NewClient(
		cfg.Redis.Addr(),
		cfg.Redis.Password,
		cfg.Redis.DB,
		l,
	)
	defer client.Close()

	// ---------------------------------------------------------
	// CONFIGURATION: Define the scope of data generation.
	// ---------------------------------------------------------
	payload := asynq.SeedPayload{
		UsersCount:           100,
		RolesCount:           10,
		PermissionsCount:     20,
		PoliciesCount:        20,
		AnnouncementsCount:   10,
		NotificationsCount:   30,
		FeatureFlagsCount:    15,
		IntegrationsCount:    5,
		TranslationsCount:    50,
		FileMetadataCount:    20,
		SiteSettingsCount:    15,
		ErrorCodesCount:      20,
		IPRulesCount:         10,
		RateLimitsCount:      8,
		AuditLogsCount:       50,
		FunctionMetricsCount: 30,
		Seed:                 0,
		ClearData:            true,
	}
	// ---------------------------------------------------------

	l.Infoc(ctx, "Enqueuing seed task...",
		zap.Int("users", payload.UsersCount),
		zap.Bool("clear_data", payload.ClearData),
	)

	// 4. Dispatch the seeding task to the queue.
	// The actual heavy lifting is performed by an Asynq worker process.
	info, err := client.EnqueueSeed(ctx, payload)
	if err != nil {
		l.Fatalc(ctx, "failed to enqueue seed task", zap.Error(err))
	}

	l.Infoc(ctx, "Seeding task enqueued successfully",
		zap.String("task_id", info.ID),
		zap.String("queue", info.Queue),
	)
}
