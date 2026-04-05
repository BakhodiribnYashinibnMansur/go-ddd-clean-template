package featureflag

import (
	"context"

	ffcache "gct/internal/context/admin/generic/featureflag/infrastructure/cache"
	"gct/internal/context/admin/generic/featureflag/infrastructure/postgres"

	"gct/internal/context/admin/generic/featureflag/application/command"
	"gct/internal/context/admin/generic/featureflag/application/query"
	"gct/internal/kernel/application"
	shareddomain "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the FeatureFlag BC.
type BoundedContext struct {
	// Commands
	CreateFlag      *command.CreateHandler
	UpdateFlag      *command.UpdateHandler
	DeleteFlag      *command.DeleteHandler
	CreateRuleGroup *command.CreateRuleGroupHandler
	UpdateRuleGroup *command.UpdateRuleGroupHandler
	DeleteRuleGroup *command.DeleteRuleGroupHandler

	// Queries
	GetFlag           *query.GetHandler
	ListFlags         *query.ListHandler
	EvaluateFlag      *query.EvaluateHandler
	BatchEvaluateFlag *query.BatchEvaluateHandler

	// Evaluator (cached, implements domain.Evaluator)
	Evaluator *ffcache.CachedEvaluator
}

// NewBoundedContext creates a fully wired FeatureFlag bounded context.
func NewBoundedContext(ctx context.Context, pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) (*BoundedContext, error) {
	writeRepo := postgres.NewFeatureFlagWriteRepo(pool)
	readRepo := postgres.NewFeatureFlagReadRepo(pool)
	rgRepo := postgres.NewRuleGroupWriteRepo(pool)

	cachedEval, err := ffcache.NewCachedEvaluator(ctx, writeRepo, l)
	if err != nil {
		return nil, err
	}

	// Subscribe to domain events to invalidate the cache.
	for _, eventName := range []string{
		"featureflag.created",
		"featureflag.updated",
		"featureflag.deleted",
		"featureflag.toggled",
	} {
		_ = eventBus.Subscribe(eventName, func(ctx context.Context, event shareddomain.DomainEvent) error {
			cachedEval.Invalidate(ctx)
			return nil
		})
	}

	return &BoundedContext{
		CreateFlag:      command.NewCreateHandler(writeRepo, eventBus, l),
		UpdateFlag:      command.NewUpdateHandler(writeRepo, eventBus, l),
		DeleteFlag:      command.NewDeleteHandler(writeRepo, eventBus, l),
		CreateRuleGroup: command.NewCreateRuleGroupHandler(writeRepo, rgRepo, eventBus, l),
		UpdateRuleGroup: command.NewUpdateRuleGroupHandler(rgRepo, eventBus, l),
		DeleteRuleGroup: command.NewDeleteRuleGroupHandler(rgRepo, eventBus, l),
		GetFlag:           query.NewGetHandler(readRepo, l),
		ListFlags:         query.NewListHandler(readRepo, l),
		EvaluateFlag:      query.NewEvaluateHandler(cachedEval),
		BatchEvaluateFlag: query.NewBatchEvaluateHandler(cachedEval),
		Evaluator:         cachedEval,
	}, nil
}
