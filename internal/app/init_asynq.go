package app

import (
	"context"
	"encoding/json"
	"fmt"

	"gct/config"
	auditcmd "gct/internal/audit/application/command"
	"gct/internal/seeder"
	"gct/internal/shared/infrastructure/asynq"
	"gct/internal/shared/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	hibikenasynq "github.com/hibiken/asynq"
)

func initAsynqWorker(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool, createAuditLog *auditcmd.CreateAuditLogHandler, l logger.Log) (*asynq.Worker, error) {
	if !cfg.Asynq.WorkerEnabled {
		l.Infoc(ctx, "⚠️ Asynq worker is disabled via configuration")
		return nil, nil
	}

	// Cluster configuration for the worker.
	if cfg.Asynq.RedisAddr == "" {
		if cfg.Database.Redis.Enabled {
			cfg.Asynq.RedisAddr = cfg.Redis.Addr()
			cfg.Asynq.RedisPassword = cfg.Redis.Password
			cfg.Asynq.RedisDB = cfg.Redis.DB
		} else {
			l.Warnc(ctx, "⚠️ Asynq worker enabled but Redis is disabled. Worker may fail to connect if no address provided.")
			// Trying to fallback or just proceeding, will likely fail in NewWorker if address is empty
		}
	}

	asynqWorker := asynq.NewWorker(cfg.Asynq, l)

	// Setup task handlers (Email, Notifications, Image processing).
	handlers := asynq.NewHandlers(l, createAuditLog)
	asynqWorker.RegisterHandler(asynq.TypeImageResize, handlers.HandleImageResize)
	asynqWorker.RegisterHandler(asynq.TypePushNotification, handlers.HandlePushNotification)
	asynqWorker.RegisterHandler(asynq.TypeAuditLog, handlers.HandleAuditLog)

	// specialized handler for on-demand database seeding.
	s := seeder.New(pool, l, cfg)
	asynqWorker.RegisterHandler(asynq.TypeSystemSeed, func(ctx context.Context, task *hibikenasynq.Task) error {
		var payload asynq.SeedPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("unmarshal seed payload: %w", err)
		}

		customCounts := make(map[string]int)
		if payload.UsersCount > 0 {
			customCounts["users"] = payload.UsersCount
		}
		if payload.RolesCount > 0 {
			customCounts["roles"] = payload.RolesCount
		}
		if payload.PermissionsCount > 0 {
			customCounts["permissions"] = payload.PermissionsCount
		}
		if payload.PoliciesCount > 0 {
			customCounts["policies"] = payload.PoliciesCount
		}
		if payload.Seed != 0 {
			customCounts["seed"] = int(payload.Seed)
		}
		customCounts["clear_data"] = 0
		if payload.ClearData {
			customCounts["clear_data"] = 1
		}

		return s.Seed(ctx, customCounts)
	})

	// Launch the worker engine in a managed goroutine.
	go func() {
		if err := asynqWorker.Start(); err != nil {
			l.Errorc(ctx, "❌ Failed to start Asynq worker",
				"error", err,
				"redis_addr", cfg.Asynq.RedisAddr,
				"concurrency", cfg.Asynq.Concurrency,
			)
		}
	}()

	return asynqWorker, nil
}
