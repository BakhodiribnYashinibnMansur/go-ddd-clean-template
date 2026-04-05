// Package asynq holds the audit BC's own asynq task handler. By keeping it
// inside the bounded context we avoid platform/ reaching into a BC. The
// composition root wires this handler into the asynq worker mux.
package asynq

import (
	"context"
	"encoding/json"
	"fmt"

	auditcmd "gct/internal/context/iam/audit/application/command"
	auditdomain "gct/internal/context/iam/audit/domain"
	"gct/internal/kernel/infrastructure/logger"

	hibikenasynq "github.com/hibiken/asynq"
	"github.com/google/uuid"
)

// TaskType is the asynq task name used for deferred audit log persistence.
const TaskType = "audit:log"

// Payload carries the audit fields for a deferred write via the task queue.
type Payload struct {
	UserID       *uuid.UUID        `json:"user_id,omitempty"`
	SessionID    *uuid.UUID        `json:"session_id,omitempty"`
	Action       string            `json:"action"`
	ResourceType *string           `json:"resource_type,omitempty"`
	ResourceID   *uuid.UUID        `json:"resource_id,omitempty"`
	Platform     *string           `json:"platform,omitempty"`
	IPAddress    *string           `json:"ip_address,omitempty"`
	UserAgent    *string           `json:"user_agent,omitempty"`
	Permission   *string           `json:"permission,omitempty"`
	PolicyID     *uuid.UUID        `json:"policy_id,omitempty"`
	Decision     *string           `json:"decision,omitempty"`
	Success      bool              `json:"success"`
	ErrorMessage *string           `json:"error_message,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// TaskHandler consumes audit log tasks and delegates to the audit BC's
// application command. It implements the asynq handler signature expected by
// the shared worker.
type TaskHandler struct {
	log    logger.Log
	create *auditcmd.CreateAuditLogHandler
}

// NewTaskHandler builds the audit asynq task handler.
func NewTaskHandler(log logger.Log, create *auditcmd.CreateAuditLogHandler) *TaskHandler {
	return &TaskHandler{log: log, create: create}
}

// Handle processes an audit log task.
func (h *TaskHandler) Handle(ctx context.Context, task *hibikenasynq.Task) error {
	var p Payload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		h.log.Errorc(ctx, "failed to unmarshal audit payload", "error", err)
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	cmd := auditcmd.CreateAuditLogCommand{
		UserID:       p.UserID,
		SessionID:    p.SessionID,
		Action:       auditdomain.AuditAction(p.Action),
		ResourceType: p.ResourceType,
		ResourceID:   p.ResourceID,
		Platform:     p.Platform,
		IPAddress:    p.IPAddress,
		UserAgent:    p.UserAgent,
		Permission:   p.Permission,
		PolicyID:     p.PolicyID,
		Decision:     p.Decision,
		Success:      p.Success,
		ErrorMessage: p.ErrorMessage,
		Metadata:     p.Metadata,
	}

	if err := h.create.Handle(ctx, cmd); err != nil {
		h.log.Errorc(ctx, "failed to save audit log via task", "error", err)
		return err
	}
	return nil
}
