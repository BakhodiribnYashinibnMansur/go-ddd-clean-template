package usecase

import (
	"context"
	"time"

	"gct/internal/domain"
	"gct/internal/shared/infrastructure/asynq"

	"github.com/google/uuid"
)

// LogAction records a specific business action to the audit log.
func (u *UseCase) LogAction(ctx context.Context, action domain.AuditActionType, userID *uuid.UUID, resourceType string, resourceID *uuid.UUID, metadata map[string]any) {
	al := &domain.AuditLog{
		ID:           uuid.New(),
		UserID:       userID,
		Action:       action,
		ResourceType: &resourceType,
		ResourceID:   resourceID,
		Metadata:     metadata,
		Success:      true,
		CreatedAt:    time.Now(),
	}

	payload := asynq.AuditPayload{
		UserID:       al.UserID,
		Action:       string(al.Action),
		ResourceType: al.ResourceType,
		ResourceID:   al.ResourceID,
	}

	// Reliable Audit Logging using Asynq
	if u.AsynqClient != nil {
		_, err := u.AsynqClient.EnqueueAudit(ctx, payload)
		if err != nil {
			u.Audit.Log().Create(ctx, al) // Fallback to direct call if enqueue fails
		}
		return
	}

	// Fallback to async direct call if AsynqClient is nil
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = u.Audit.Log().Create(bgCtx, al)
	}()
}
