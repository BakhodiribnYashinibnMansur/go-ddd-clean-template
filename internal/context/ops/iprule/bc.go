package iprule

import (
	"gct/internal/context/ops/iprule/application/command"
	"gct/internal/context/ops/iprule/application/query"
	"gct/internal/context/ops/iprule/infrastructure/postgres"
	"gct/internal/kernel/application"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the IPRule BC.
type BoundedContext struct {
	// Commands
	CreateIPRule *command.CreateIPRuleHandler
	UpdateIPRule *command.UpdateIPRuleHandler
	DeleteIPRule *command.DeleteIPRuleHandler

	// Queries
	GetIPRule   *query.GetIPRuleHandler
	ListIPRules *query.ListIPRulesHandler
}

// NewBoundedContext creates a fully wired IPRule bounded context.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewIPRuleWriteRepo(pool)
	readRepo := postgres.NewIPRuleReadRepo(pool)

	return &BoundedContext{
		CreateIPRule: command.NewCreateIPRuleHandler(writeRepo, eventBus, l),
		UpdateIPRule: command.NewUpdateIPRuleHandler(writeRepo, eventBus, l),
		DeleteIPRule: command.NewDeleteIPRuleHandler(writeRepo, l),
		GetIPRule:    query.NewGetIPRuleHandler(readRepo, l),
		ListIPRules:  query.NewListIPRulesHandler(readRepo, l),
	}
}
