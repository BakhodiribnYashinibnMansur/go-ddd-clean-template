package query

import (
	"context"

	appdto "gct/internal/job/application"
	"gct/internal/job/domain"
)

// ListJobsQuery holds the input for listing jobs with filtering.
type ListJobsQuery struct {
	Filter domain.JobFilter
}

// ListJobsResult holds the output of the list jobs query.
type ListJobsResult struct {
	Jobs  []*appdto.JobView
	Total int64
}

// ListJobsHandler handles the ListJobsQuery.
type ListJobsHandler struct {
	readRepo domain.JobReadRepository
}

// NewListJobsHandler creates a new ListJobsHandler.
func NewListJobsHandler(readRepo domain.JobReadRepository) *ListJobsHandler {
	return &ListJobsHandler{readRepo: readRepo}
}

// Handle executes the ListJobsQuery and returns a list of JobView with total count.
func (h *ListJobsHandler) Handle(ctx context.Context, q ListJobsQuery) (*ListJobsResult, error) {
	views, total, err := h.readRepo.List(ctx, q.Filter)
	if err != nil {
		return nil, err
	}

	result := make([]*appdto.JobView, len(views))
	for i, v := range views {
		result[i] = &appdto.JobView{
			ID:          v.ID,
			TaskName:    v.TaskName,
			Status:      v.Status,
			Payload:     v.Payload,
			Result:      v.Result,
			Attempts:    v.Attempts,
			MaxAttempts: v.MaxAttempts,
			ScheduledAt: v.ScheduledAt,
			StartedAt:   v.StartedAt,
			CompletedAt: v.CompletedAt,
			Error:       v.Error,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		}
	}

	return &ListJobsResult{
		Jobs:  result,
		Total: total,
	}, nil
}
