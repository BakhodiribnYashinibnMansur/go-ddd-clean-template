package errors

import (
	"context"
	"sync"
	"time"

	"gct/internal/shared/infrastructure/contextx"
)

// ChainEntry represents one error in a request's error chain.
type ChainEntry struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Layer     string    `json:"layer"`
	Timestamp time.Time `json:"timestamp"`
}

// ErrorChain tracks errors grouped by request_id.
type ErrorChain struct {
	mu     sync.RWMutex
	chains map[string][]ChainEntry // request_id → entries
	maxAge time.Duration
}

// NewErrorChain creates a new error chain tracker.
func NewErrorChain(maxAge time.Duration) *ErrorChain {
	if maxAge <= 0 {
		maxAge = 5 * time.Minute
	}
	return &ErrorChain{
		chains: make(map[string][]ChainEntry),
		maxAge: maxAge,
	}
}

// Record adds an error to the chain for the current request.
func (c *ErrorChain) Record(ctx context.Context, err *AppError) {
	reqID := contextx.GetRequestID(ctx)
	if reqID == "" || err == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.chains[reqID] = append(c.chains[reqID], ChainEntry{
		Code:      err.Type,
		Message:   err.Message,
		Layer:     string(GetLayer(err.Type)),
		Timestamp: time.Now(),
	})
}

// Get returns the error chain for a request_id.
func (c *ErrorChain) Get(requestID string) []ChainEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entries := c.chains[requestID]
	result := make([]ChainEntry, len(entries))
	copy(result, entries)
	return result
}

// Cleanup removes expired chains.
func (c *ErrorChain) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for reqID, entries := range c.chains {
		if len(entries) > 0 && now.Sub(entries[0].Timestamp) > c.maxAge {
			delete(c.chains, reqID)
		}
	}
}

// StartCleanup runs periodic cleanup.
func (c *ErrorChain) StartCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(c.maxAge)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.Cleanup()
			}
		}
	}()
}

// ChainHook returns an ErrorHook that records errors into the chain.
func ChainHook(chain *ErrorChain) func(ctx context.Context, err *AppError) {
	return func(ctx context.Context, err *AppError) {
		chain.Record(ctx, err)
	}
}
