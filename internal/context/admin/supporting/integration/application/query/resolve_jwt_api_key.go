package query

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"sync"
	"time"

	integentity "gct/internal/context/admin/supporting/integration/domain/entity"
	integrepo "gct/internal/context/admin/supporting/integration/domain/repository"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// ResolveJWTAPIKeyQuery carries the plaintext X-API-Key header value.
type ResolveJWTAPIKeyQuery struct {
	PlainAPIKey string
}

// resolveCacheEntry holds a cached JWTIntegrationView and its insertion time.
type resolveCacheEntry struct {
	view       *integentity.JWTIntegrationView
	insertedAt time.Time
}

// ResolveJWTAPIKeyHandler resolves a plaintext JWT API key against the
// integration store using HMAC-SHA256(pepper, plaintext). Results are cached
// in-process for cacheTTL to avoid a DB round-trip per sign-in attempt.
type ResolveJWTAPIKeyHandler struct {
	readRepo integrepo.IntegrationReadRepository
	pepper   []byte
	cacheTTL time.Duration
	logger   logger.Log

	mu    sync.RWMutex
	cache map[string]resolveCacheEntry
}

// NewResolveJWTAPIKeyHandler constructs a ResolveJWTAPIKeyHandler.
func NewResolveJWTAPIKeyHandler(repo integrepo.IntegrationReadRepository, pepper []byte, cacheTTL time.Duration, l logger.Log) *ResolveJWTAPIKeyHandler {
	return &ResolveJWTAPIKeyHandler{
		readRepo: repo,
		pepper:   pepper,
		cacheTTL: cacheTTL,
		logger:   l,
		cache:    make(map[string]resolveCacheEntry),
	}
}

// cacheKey is the in-process cache key: HMAC(pepper, plaintext) as a string.
// It is stable for the life of the process and leaks nothing beyond the hash.
func (h *ResolveJWTAPIKeyHandler) hash(plaintext string) []byte {
	mac := hmac.New(sha256.New, h.pepper)
	mac.Write([]byte(plaintext))
	return mac.Sum(nil)
}

// Handle resolves a plaintext API key to its JWTIntegrationView.
// Returns integentity.ErrInvalidJWTAPIKey if the key is too short, or a mapped
// service error (ErrAPIKeyNotFound) if no integration matches the hash.
func (h *ResolveJWTAPIKeyHandler) Handle(ctx context.Context, q ResolveJWTAPIKeyQuery) (result *integentity.JWTIntegrationView, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ResolveJWTAPIKeyHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ResolveJWTAPIKey", "integration")()

	if len(q.PlainAPIKey) < 32 {
		return nil, integentity.ErrInvalidJWTAPIKey
	}

	hashed := h.hash(q.PlainAPIKey)
	key := string(hashed)

	// Cache read path.
	h.mu.RLock()
	entry, ok := h.cache[key]
	h.mu.RUnlock()
	if ok {
		if h.cacheTTL > 0 && time.Since(entry.insertedAt) < h.cacheTTL {
			return entry.view, nil
		}
		// Evict stale entry.
		h.mu.Lock()
		delete(h.cache, key)
		h.mu.Unlock()
	}

	view, err := h.readRepo.FindJWTByHash(ctx, hashed)
	if err != nil {
		if errors.Is(err, integentity.ErrIntegrationNotFound) {
			return nil, integentity.ErrAPIKeyNotFound
		}
		return nil, apperrors.MapToServiceError(err)
	}

	h.mu.Lock()
	h.cache[key] = resolveCacheEntry{view: view, insertedAt: time.Now()}
	h.mu.Unlock()

	return view, nil
}

// Invalidate drops a cached entry for the given plaintext API key. Callers
// should invoke this when the underlying integration has been updated.
func (h *ResolveJWTAPIKeyHandler) Invalidate(plainAPIKey string) {
	key := string(h.hash(plainAPIKey))
	h.mu.Lock()
	delete(h.cache, key)
	h.mu.Unlock()
}
