package csrf

import (
	"context"
	"fmt"
)

// RotationManager handles session and CSRF token rotation
type RotationManager struct {
	generator *Generator
	store     Store
}

// NewRotationManager creates a new rotation manager
func NewRotationManager(generator *Generator, store Store) *RotationManager {
	return &RotationManager{
		generator: generator,
		store:     store,
	}
}

// RotateOnLogin rotates both session and CSRF token after successful login
// This prevents session fixation attacks
func (rm *RotationManager) RotateOnLogin(ctx context.Context, oldSessionID, newSessionID string) (*Token, error) {
	// Invalidate old CSRF token
	if oldSessionID != "" {
		_ = rm.store.Delete(ctx, oldSessionID)
	}

	// Generate new CSRF token for new session
	newToken, err := rm.generator.GenerateToken(newSessionID)
	if err != nil {
		return nil, err
	}

	// Store new token hash
	if err := rm.store.Set(ctx, newSessionID, newToken.Hash, rm.generator.expiration); err != nil {
		return nil, fmt.Errorf("csrf.RotationManager.store.Set: %w", err)
	}

	return newToken, nil
}

// RotateOnPasswordChange rotates CSRF token after password change
// This is critical for security after credential changes
func (rm *RotationManager) RotateOnPasswordChange(ctx context.Context, sessionID string) (*Token, error) {
	return rm.rotateToken(ctx, sessionID)
}

// RotateOnPrivilegeChange rotates CSRF token after role/permission changes
// Prevents privilege escalation attacks
func (rm *RotationManager) RotateOnPrivilegeChange(ctx context.Context, sessionID string) (*Token, error) {
	return rm.rotateToken(ctx, sessionID)
}

// RotateOnRefresh rotates CSRF token during token refresh
// Provides additional security for long-lived sessions
func (rm *RotationManager) RotateOnRefresh(ctx context.Context, sessionID string) (*Token, error) {
	return rm.rotateToken(ctx, sessionID)
}

// InvalidateOnLogout invalidates CSRF token on logout
// Note: Session should also be invalidated separately
func (rm *RotationManager) InvalidateOnLogout(ctx context.Context, sessionID string) error {
	if err := rm.store.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("csrf.RotationManager.store.Delete: %w", err)
	}
	return nil
}

// rotateToken is the internal rotation logic
func (rm *RotationManager) rotateToken(ctx context.Context, sessionID string) (*Token, error) {
	// Generate new token
	newToken, err := rm.generator.GenerateToken(sessionID)
	if err != nil {
		return nil, err
	}

	// Replace old token with new one atomically
	if err := rm.store.Rotate(ctx, sessionID, newToken.Hash, rm.generator.expiration); err != nil {
		return nil, fmt.Errorf("csrf.RotationManager.store.Rotate: %w", err)
	}

	return newToken, nil
}
