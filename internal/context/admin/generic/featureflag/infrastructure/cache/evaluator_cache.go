package cache

import (
	"context"
	"strconv"
	"sync"

	"gct/internal/context/admin/generic/featureflag/application/query"
	"gct/internal/context/admin/generic/featureflag/domain"
	"gct/internal/kernel/infrastructure/logger"
)

// CachedEvaluator evaluates feature flags from an in-memory cache backed by
// the FeatureFlagRepository. On startup all flags are loaded; the cache can be
// invalidated (reloaded) via the Invalidate method.
type CachedEvaluator struct {
	repo    domain.FeatureFlagRepository
	cache   sync.Map
	loadMu  sync.Mutex // serializes LoadAll to prevent interleaved clear+populate
	log     logger.Log
}

// NewCachedEvaluator creates a CachedEvaluator and eagerly loads all flags.
func NewCachedEvaluator(ctx context.Context, repo domain.FeatureFlagRepository, log logger.Log) (*CachedEvaluator, error) {
	ce := &CachedEvaluator{repo: repo, log: log}
	if err := ce.LoadAll(ctx); err != nil {
		return nil, err
	}
	return ce, nil
}

// LoadAll fetches every flag from the repository and atomically replaces the cache.
func (ce *CachedEvaluator) LoadAll(ctx context.Context) error {
	ce.loadMu.Lock()
	defer ce.loadMu.Unlock()

	flags, err := ce.repo.FindAll(ctx)
	if err != nil {
		return err
	}
	ce.cache.Range(func(key, _ any) bool { ce.cache.Delete(key); return true })
	for _, ff := range flags {
		ce.cache.Store(ff.Key(), ff)
	}
	ce.log.Infow("feature flag cache loaded", "count", len(flags))
	return nil
}

// Invalidate reloads the entire cache from the repository.
func (ce *CachedEvaluator) Invalidate(ctx context.Context) {
	if err := ce.LoadAll(ctx); err != nil {
		ce.log.Errorw("failed to reload feature flag cache", "error", err)
	}
}

func (ce *CachedEvaluator) IsEnabled(ctx context.Context, flagKey string, userAttrs map[string]string) bool {
	ff := ce.getFlag(ctx, flagKey)
	if ff == nil {
		return false
	}
	return ff.Evaluate(userAttrs) == "true"
}

func (ce *CachedEvaluator) GetString(ctx context.Context, flagKey string, userAttrs map[string]string) string {
	ff := ce.getFlag(ctx, flagKey)
	if ff == nil {
		return ""
	}
	return ff.Evaluate(userAttrs)
}

func (ce *CachedEvaluator) GetInt(ctx context.Context, flagKey string, userAttrs map[string]string) int {
	ff := ce.getFlag(ctx, flagKey)
	if ff == nil {
		return 0
	}
	val, err := strconv.Atoi(ff.Evaluate(userAttrs))
	if err != nil {
		return 0
	}
	return val
}

func (ce *CachedEvaluator) GetFloat(ctx context.Context, flagKey string, userAttrs map[string]string) float64 {
	ff := ce.getFlag(ctx, flagKey)
	if ff == nil {
		return 0
	}
	val, err := strconv.ParseFloat(ff.Evaluate(userAttrs), 64)
	if err != nil {
		return 0
	}
	return val
}

// EvaluateFull evaluates a flag and returns the value together with its type.
// Returns nil when the flag does not exist.
func (ce *CachedEvaluator) EvaluateFull(ctx context.Context, key string, userAttrs map[string]string) *query.EvalResult {
	ff := ce.getFlag(ctx, key)
	if ff == nil {
		return nil
	}
	return &query.EvalResult{
		Value:    ff.Evaluate(userAttrs),
		FlagType: ff.FlagType(),
	}
}

func (ce *CachedEvaluator) getFlag(ctx context.Context, key string) *domain.FeatureFlag {
	if val, ok := ce.cache.Load(key); ok {
		return val.(*domain.FeatureFlag)
	}
	ff, err := ce.repo.FindByKey(ctx, key)
	if err != nil {
		ce.log.Debugw("feature flag not found", "key", key)
		return nil
	}
	ce.cache.Store(key, ff)
	return ff
}
