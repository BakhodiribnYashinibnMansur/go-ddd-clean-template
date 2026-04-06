package postgres

import (
	"context"
	"strconv"

	ffrepo "gct/internal/context/admin/generic/featureflag/domain/repository"
)

// PgEvaluator evaluates feature flags by querying PostgreSQL on every call.
type PgEvaluator struct {
	repo ffrepo.FeatureFlagRepository
}

// NewPgEvaluator creates a new PostgreSQL-backed Evaluator.
func NewPgEvaluator(repo ffrepo.FeatureFlagRepository) *PgEvaluator {
	return &PgEvaluator{repo: repo}
}

func (e *PgEvaluator) IsEnabled(ctx context.Context, flagKey string, userAttrs map[string]string) bool {
	ff, err := e.repo.FindByKey(ctx, flagKey)
	if err != nil {
		return false
	}
	return ff.Evaluate(userAttrs) == "true"
}

func (e *PgEvaluator) GetString(ctx context.Context, flagKey string, userAttrs map[string]string) string {
	ff, err := e.repo.FindByKey(ctx, flagKey)
	if err != nil {
		return ""
	}
	return ff.Evaluate(userAttrs)
}

func (e *PgEvaluator) GetInt(ctx context.Context, flagKey string, userAttrs map[string]string) int {
	ff, err := e.repo.FindByKey(ctx, flagKey)
	if err != nil {
		return 0
	}
	val, err := strconv.Atoi(ff.Evaluate(userAttrs))
	if err != nil {
		return 0
	}
	return val
}

func (e *PgEvaluator) GetFloat(ctx context.Context, flagKey string, userAttrs map[string]string) float64 {
	ff, err := e.repo.FindByKey(ctx, flagKey)
	if err != nil {
		return 0
	}
	val, err := strconv.ParseFloat(ff.Evaluate(userAttrs), 64)
	if err != nil {
		return 0
	}
	return val
}
