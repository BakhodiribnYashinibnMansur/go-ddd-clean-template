package asynq

import (
	"context"
	"encoding/json"
	"fmt"

	auditcmd "gct/internal/audit/application/command"
	auditdomain "gct/internal/audit/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/google/uuid"
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
	PoliciesCount    int   `json:"policies_count"`
	Seed             int64 `json:"seed"`
	ClearData        bool  `json:"clear_data"`
}

// AuditPayload represents audit log task payload.
type AuditPayload struct {
	UserID       *uuid.UUID     `json:"user_id,omitempty"`
	SessionID    *uuid.UUID     `json:"session_id,omitempty"`
	Action       string         `json:"action"`
	ResourceType *string        `json:"resource_type,omitempty"`
	ResourceID   *uuid.UUID     `json:"resource_id,omitempty"`
	Platform     *string        `json:"platform,omitempty"`
	IPAddress    *string        `json:"ip_address,omitempty"`
	UserAgent    *string        `json:"user_agent,omitempty"`
	Permission   *string        `json:"permission,omitempty"`
	PolicyID     *uuid.UUID     `json:"policy_id,omitempty"`
	Decision     *string        `json:"decision,omitempty"`
	Success      bool           `json:"success"`
	ErrorMessage *string        `json:"error_message,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Handlers contains all task handlers.
type Handlers struct {
	log            logger.Log
	createAuditLog *auditcmd.CreateAuditLogHandler
}

// NewHandlers creates a new handlers instance.
func NewHandlers(log logger.Log, createAuditLog *auditcmd.CreateAuditLogHandler) *Handlers {
	return &Handlers{log: log, createAuditLog: createAuditLog}
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

// HandleAuditLog processes audit log task.
func (h *Handlers) HandleAuditLog(ctx context.Context, task *asynq.Task) error {
	var payload AuditPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		h.log.Errorc(ctx, "failed to unmarshal audit payload", "error", err)
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	cmd := auditcmd.CreateAuditLogCommand{
		UserID:       payload.UserID,
		SessionID:    payload.SessionID,
		Action:       auditdomain.AuditAction(payload.Action),
		ResourceType: payload.ResourceType,
		ResourceID:   payload.ResourceID,
		Platform:     payload.Platform,
		IPAddress:    payload.IPAddress,
		UserAgent:    payload.UserAgent,
		Permission:   payload.Permission,
		PolicyID:     payload.PolicyID,
		Decision:     payload.Decision,
		Success:      payload.Success,
		ErrorMessage: payload.ErrorMessage,
		Metadata:     payload.Metadata,
	}

	if err := h.createAuditLog.Handle(ctx, cmd); err != nil {
		h.log.Errorc(ctx, "failed to save audit log via task", "error", err)
		return err
	}

	return nil
}
