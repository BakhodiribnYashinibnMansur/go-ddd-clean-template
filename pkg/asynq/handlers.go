package asynq

import (
	"context"
	"encoding/json"
	"fmt"

	"gct/internal/domain"
	"gct/internal/usecase/audit"
	"gct/pkg/logger"

	"github.com/hibiken/asynq"
)

// EmailPayload represents email task payload.
type EmailPayload struct {
	To      string            `json:"to"`
	Subject string            `json:"subject"`
	Body    string            `json:"body"`
	Data    map[string]string `json:"data,omitempty"`
}

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
	PoliciesCount    int   `json:"policies_count"`
	Seed             int64 `json:"seed"`
	ClearData        bool  `json:"clear_data"`
}

// AuditPayload represents audit log task payload.
type AuditPayload struct {
	Log *domain.AuditLog `json:"log"`
}

// Handlers contains all task handlers.
type Handlers struct {
	log   logger.Log
	audit audit.UseCaseI
}

// NewHandlers creates a new handlers instance.
func NewHandlers(log logger.Log, audit audit.UseCaseI) *Handlers {
	return &Handlers{log: log, audit: audit}
}

// HandleEmailWelcome processes welcome email task.
func (h *Handlers) HandleEmailWelcome(ctx context.Context, task *asynq.Task) error {
	var payload EmailPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		h.log.Errorc(ctx, "failed to unmarshal email payload", "error", err)
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	h.log.Infoc(ctx, "processing welcome email",
		"to", payload.To,
		"subject", payload.Subject,
	)

	// TODO: Implement actual email sending logic
	// Example: emailService.Send(ctx, payload)

	h.log.Infoc(ctx, "welcome email sent successfully",
		"to", payload.To,
	)

	return nil
}

// HandleEmailVerification processes email verification task.
func (h *Handlers) HandleEmailVerification(ctx context.Context, task *asynq.Task) error {
	var payload EmailPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		h.log.Errorc(ctx, "failed to unmarshal email payload", "error", err)
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	h.log.Infoc(ctx, "processing verification email",
		"to", payload.To,
	)

	// TODO: Implement actual email sending logic

	h.log.Infoc(ctx, "verification email sent successfully",
		"to", payload.To,
	)

	return nil
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

	// TODO: Implement actual image resizing logic
	// Example: imageService.Resize(ctx, payload)

	h.log.Infoc(ctx, "image resized successfully",
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

	// TODO: Implement actual push notification logic
	// Example: notificationService.SendPush(ctx, payload)

	h.log.Infoc(ctx, "push notification sent successfully",
		"user_id", payload.UserID,
	)

	return nil
}

// HandleAuditLog processes audit log task.
func (h *Handlers) HandleAuditLog(ctx context.Context, task *asynq.Task) error {
	var payload AuditPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		h.log.Errorc(ctx, "failed to unmarshal audit payload", "error", err)
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	if err := h.audit.Log().Create(ctx, payload.Log); err != nil {
		h.log.Errorc(ctx, "failed to save audit log via task", "error", err)
		return err
	}

	return nil
}
