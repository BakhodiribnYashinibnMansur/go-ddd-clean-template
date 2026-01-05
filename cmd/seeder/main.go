package main

import (
	"context"
	"fmt"
	"gct/config"
	"gct/pkg/asynq"
	"gct/pkg/logger"
	"os"

	"go.uber.org/zap"
)

func main() {
	// Create context
	ctx := context.Background()

	// Load config
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	l := logger.New(cfg.Log.Level)

	// Initialize Asynq Client
	client := asynq.NewClient(
		cfg.Redis.Addr(),
		cfg.Redis.Password,
		cfg.Redis.DB,
		l,
	)
	defer client.Close()

	// ---------------------------------------------------------
	// CONFIGURATION: Set your seeding parameters here
	// ---------------------------------------------------------
	payload := asynq.SeedPayload{
		UsersCount:       100,  // Nechta user yaratish
		RolesCount:       10,   // Nechta rol yaratish
		PermissionsCount: 20,   // Nechta permission yaratish
		PoliciesCount:    20,   // Nechta policy yaratish
		Seed:             0,    // 0 = random, boshqa son = reproducibility
		ClearData:        true, // Bazani tozalash kerakmi?
	}
	// ---------------------------------------------------------

	l.WithContext(ctx).Infow("Enqueuing seed task...",
		zap.Int("users", payload.UsersCount),
		zap.Bool("clear_data", payload.ClearData),
	)

	// Enqueue the task
	info, err := client.EnqueueSeed(ctx, payload)
	if err != nil {
		l.WithContext(ctx).Fatalw("failed to enqueue seed task", zap.Error(err))
	}

	l.WithContext(ctx).Infow("Seeding task enqueued successfully",
		zap.String("task_id", info.ID),
		zap.String("queue", info.Queue),
	)
}
