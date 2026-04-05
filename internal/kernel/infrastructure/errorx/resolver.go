package errorx

import (
	"context"
	"sync"
)

// ResolutionAction is a function that attempts to resolve/mitigate an error.
// Returns true if the resolution was applied.
type ResolutionAction func(ctx context.Context, err *AppError) bool

// Resolver maps error codes to automatic resolution actions.
type Resolver struct {
	mu      sync.RWMutex
	actions map[string]ResolutionAction
}

// NewResolver creates a new error resolver.
func NewResolver() *Resolver {
	return &Resolver{
		actions: make(map[string]ResolutionAction),
	}
}

// Register adds a resolution action for an error code.
func (r *Resolver) Register(code string, action ResolutionAction) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.actions[code] = action
}

// Resolve attempts to automatically resolve an error.
// Returns true if a resolution was applied.
func (r *Resolver) Resolve(ctx context.Context, err *AppError) bool {
	if err == nil {
		return false
	}

	r.mu.RLock()
	action, ok := r.actions[err.Type]
	r.mu.RUnlock()

	if !ok {
		return false
	}
	return action(ctx, err)
}

// ResolverHook returns an ErrorHook that attempts auto-resolution.
func ResolverHook(resolver *Resolver) func(ctx context.Context, err *AppError) {
	return func(ctx context.Context, err *AppError) {
		resolver.Resolve(ctx, err)
	}
}
