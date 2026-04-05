package asynq

import (
	"context"
	"encoding/json"
	"fmt"

	"gct/internal/platform/infrastructure/logger"

	"github.com/hibiken/asynq"
)

// ImagePayload represents image processing task payload.
type ImagePayload struct {
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	Quality    int    `json:"quality,omitempty"`
}

// NotificationPayload represents notification task payload.
type NotificationPayload struct {
	UserID  string            `json:"user_id"`
	Title   string            `json:"title"`
	Message string            `json:"message"`
	Data    map[string]string `json:"data,omitempty"`
}

// SeedPayload represents seeding task payload.
type SeedPayload struct {
	UsersCount       int   `json:"users_count"`
	RolesCount       int   `json:"roles_count"`
	PermissionsCount int   `json:"permissions_count"`
	PoliciesCount        int   `json:"policies_count"`
	AnnouncementsCount   int   `json:"announcements_count"`
	NotificationsCount   int   `json:"notifications_count"`
	FeatureFlagsCount    int   `json:"feature_flags_count"`
	IntegrationsCount    int   `json:"integrations_count"`
	TranslationsCount    int   `json:"translations_count"`
	FileMetadataCount    int   `json:"file_metadata_count"`
	SiteSettingsCount    int   `json:"site_settings_count"`
	ErrorCodesCount      int   `json:"error_codes_count"`
	IPRulesCount         int   `json:"ip_rules_count"`
	RateLimitsCount      int   `json:"rate_limits_count"`
	AuditLogsCount       int   `json:"audit_logs_count"`
	FunctionMetricsCount int   `json:"function_metrics_count"`
	Seed                 int64 `json:"seed"`
	ClearData        bool  `json:"clear_data"`
}

// Handlers contains platform-level task handlers (image resize, push
// notifications). BC-owned handlers live inside their bounded context.
type Handlers struct {
	log logger.Log
}

// NewHandlers creates a new handlers instance.
func NewHandlers(log logger.Log) *Handlers {
	return &Handlers{log: log}
}

// HandleImageResize processes image resize task.
func (h *Handlers) HandleImageResize(ctx context.Context, task *asynq.Task) error {
	var payload ImagePayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		h.log.Errorc(ctx, "failed to unmarshal image payload", "error", err)
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	h.log.Infoc(ctx, "processing image resize",
		"source", payload.SourcePath,
		"target", payload.TargetPath,
		"width", payload.Width,
		"height", payload.Height,
	)

	h.log.Infoc(ctx, "image resize completed (no-op)",
		"target", payload.TargetPath,
	)

	return nil
}

// HandlePushNotification processes push notification task.
func (h *Handlers) HandlePushNotification(ctx context.Context, task *asynq.Task) error {
	var payload NotificationPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		h.log.Errorc(ctx, "failed to unmarshal notification payload", "error", err)
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	h.log.Infoc(ctx, "processing push notification",
		"user_id", payload.UserID,
		"title", payload.Title,
	)

	h.log.Infoc(ctx, "push notification completed (no-op)",
		"user_id", payload.UserID,
	)

	return nil
}

