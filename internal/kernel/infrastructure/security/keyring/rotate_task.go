package keyring

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gct/internal/kernel/infrastructure/logger"

	"github.com/hibiken/asynq"
)

// TaskTypeRotateKeys is the asynq task name for the key rotation worker.
const TaskTypeRotateKeys = "keyring:rotate_keys"

// RotateKeysPayload is empty because the task is self-contained: the handler
// fetches the integration list from the DB at execution time.
type RotateKeysPayload struct{}

// NewRotateKeysTask creates an asynq.Task for the rotation worker.
func NewRotateKeysTask() (*asynq.Task, error) {
	payload, err := json.Marshal(RotateKeysPayload{})
	if err != nil {
		return nil, fmt.Errorf("keyring: marshal rotate payload: %w", err)
	}
	return asynq.NewTask(TaskTypeRotateKeys, payload), nil
}

// IntegrationLister abstracts the read-side query needed by the rotation
// handler. The concrete implementation is supplied by the Integration BC's
// read repository.
type IntegrationLister interface {
	ListActiveJWT(ctx context.Context) ([]JWTIntegrationView, error)
}

// JWTIntegrationView is the rotation handler's view of an active integration.
// It mirrors the fields the handler needs without importing the Integration
// BC's domain package.
type JWTIntegrationView struct {
	Name            string
	KeyID           string
	RotatedAt       *time.Time
	RotateEveryDays int
}

// RotateKeysHandler processes the keyring:rotate_keys asynq task.
type RotateKeysHandler struct {
	kr          *Keyring
	readRepo    IntegrationLister
	writeUpdate UpdateKeyFn
	logger      logger.Log
}

// NewRotateKeysHandler builds a handler with the dependencies it needs.
func NewRotateKeysHandler(kr *Keyring, lister IntegrationLister, update UpdateKeyFn, l logger.Log) *RotateKeysHandler {
	return &RotateKeysHandler{
		kr:          kr,
		readRepo:    lister,
		writeUpdate: update,
		logger:      l,
	}
}

// Handle enumerates active integrations, checks if any are due for rotation,
// and rotates them.
func (h *RotateKeysHandler) Handle(ctx context.Context, _ *asynq.Task) error {
	views, err := h.readRepo.ListActiveJWT(ctx)
	if err != nil {
		return fmt.Errorf("keyring rotate: list active JWT integrations: %w", err)
	}

	var firstErr error
	for _, v := range views {
		if err := h.maybeRotate(ctx, v); err != nil {
			h.logger.Errorc(ctx, "keyring rotate: failed to rotate integration",
				"integration", v.Name,
				"error", err,
			)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// maybeRotate checks whether a single integration is due for rotation and
// performs the rotation if so.
func (h *RotateKeysHandler) maybeRotate(ctx context.Context, v JWTIntegrationView) error {
	// Just generated (no prior rotation) — skip.
	if v.RotatedAt == nil {
		h.logger.Infoc(ctx, "keyring rotate: skipping (never rotated, just generated)",
			"integration", v.Name,
		)
		return nil
	}

	// Not due yet.
	if v.RotateEveryDays <= 0 {
		h.logger.Infoc(ctx, "keyring rotate: skipping (rotation disabled)",
			"integration", v.Name,
		)
		return nil
	}

	rotationInterval := time.Duration(v.RotateEveryDays) * 24 * time.Hour
	if time.Since(*v.RotatedAt) < rotationInterval {
		h.logger.Infoc(ctx, "keyring rotate: skipping (not due yet)",
			"integration", v.Name,
			"rotated_at", v.RotatedAt,
			"rotate_every_days", v.RotateEveryDays,
		)
		return nil
	}

	// Perform rotation.
	kp, err := h.kr.Rotate(v.Name)
	if err != nil {
		return fmt.Errorf("rotate %s: %w", v.Name, err)
	}

	if err := h.writeUpdate(ctx, v.Name, string(kp.PublicKeyPEM), kp.KeyID); err != nil {
		return fmt.Errorf("persist rotated key %s: %w", v.Name, err)
	}

	h.logger.Infoc(ctx, "keyring rotate: rotated key pair",
		"integration", v.Name,
		"new_kid", kp.KeyID,
	)
	return nil
}
