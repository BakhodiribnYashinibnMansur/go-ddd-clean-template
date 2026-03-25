package job

import (
	"gct/internal/job/application/command"
	"gct/internal/job/application/query"
	"gct/internal/job/infrastructure/postgres"
	"gct/internal/shared/application"
	"gct/internal/shared/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the Job BC.
type BoundedContext struct {
	// Commands
	CreateJob *command.CreateJobHandler
	UpdateJob *command.UpdateJobHandler
	DeleteJob *command.DeleteJobHandler

	// Queries
	GetJob   *query.GetJobHandler
	ListJobs *query.ListJobsHandler
}

// NewBoundedContext creates a fully wired Job bounded context.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	writeRepo := postgres.NewJobWriteRepo(pool)
	readRepo := postgres.NewJobReadRepo(pool)

	return &BoundedContext{
		CreateJob: command.NewCreateJobHandler(writeRepo, eventBus, l),
		UpdateJob: command.NewUpdateJobHandler(writeRepo, eventBus, l),
		DeleteJob: command.NewDeleteJobHandler(writeRepo, l),
		GetJob:    query.NewGetJobHandler(readRepo),
		ListJobs:  query.NewListJobsHandler(readRepo),
	}
}
