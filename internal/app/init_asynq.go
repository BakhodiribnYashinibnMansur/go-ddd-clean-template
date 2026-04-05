package app

import (
	"context"
	"encoding/json"
	"fmt"

	"gct/config"
	auditcmd "gct/internal/context/iam/audit/application/command"
	auditasynq "gct/internal/context/iam/audit/infrastructure/asynq"
	"gct/internal/app/seeder"
	"gct/internal/kernel/infrastructure/asynq"
	"gct/internal/kernel/infrastructure/asynq/tasks"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	hibikenasynq "github.com/hibiken/asynq"
)

func initAsynqWorker(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool, createAuditLog *auditcmd.CreateAuditLogHandler, l logger.Log, fcmSender tasks.FCMSender, tgSender tasks.TelegramSender) (*asynq.Worker, error) {
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
	handlers := asynq.NewHandlers(l)
	asynqWorker.RegisterHandler(asynq.TypeImageResize, handlers.HandleImageResize)
	asynqWorker.RegisterHandler(asynq.TypePushNotification, handlers.HandlePushNotification)

	// BC-owned handlers register themselves through the composition root.
	auditTaskHandler := auditasynq.NewTaskHandler(l, createAuditLog)
	asynqWorker.RegisterHandler(auditasynq.TaskType, auditTaskHandler.Handle)

	// External service task handlers (Firebase, Telegram).
	asynqWorker.RegisterExternalHandlers(fcmSender, tgSender)

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
		if payload.AnnouncementsCount > 0 {
			customCounts["announcements"] = payload.AnnouncementsCount
		}
		if payload.NotificationsCount > 0 {
			customCounts["notifications"] = payload.NotificationsCount
		}
		if payload.FeatureFlagsCount > 0 {
			customCounts["feature_flags"] = payload.FeatureFlagsCount
		}
		if payload.IntegrationsCount > 0 {
			customCounts["integrations"] = payload.IntegrationsCount
		}
		if payload.TranslationsCount > 0 {
			customCounts["translations"] = payload.TranslationsCount
		}
		if payload.FileMetadataCount > 0 {
			customCounts["file_metadata"] = payload.FileMetadataCount
		}
		if payload.SiteSettingsCount > 0 {
			customCounts["site_settings"] = payload.SiteSettingsCount
		}
		if payload.ErrorCodesCount > 0 {
			customCounts["error_codes"] = payload.ErrorCodesCount
		}
		if payload.IPRulesCount > 0 {
			customCounts["ip_rules"] = payload.IPRulesCount
		}
		if payload.RateLimitsCount > 0 {
			customCounts["rate_limits"] = payload.RateLimitsCount
		}
		if payload.AuditLogsCount > 0 {
			customCounts["audit_logs"] = payload.AuditLogsCount
		}
		if payload.FunctionMetricsCount > 0 {
			customCounts["function_metrics"] = payload.FunctionMetricsCount
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
