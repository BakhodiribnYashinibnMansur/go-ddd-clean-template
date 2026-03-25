package query

import (
	"context"

	appdto "gct/internal/job/application"
	"gct/internal/job/domain"

	"github.com/google/uuid"
)

// GetJobQuery holds the input for getting a single job.
type GetJobQuery struct {
	ID uuid.UUID
}

// GetJobHandler handles the GetJobQuery.
type GetJobHandler struct {
	readRepo domain.JobReadRepository
}

// NewGetJobHandler creates a new GetJobHandler.
func NewGetJobHandler(readRepo domain.JobReadRepository) *GetJobHandler {
	return &GetJobHandler{readRepo: readRepo}
}

// Handle executes the GetJobQuery and returns a JobView.
func (h *GetJobHandler) Handle(ctx context.Context, q GetJobQuery) (*appdto.JobView, error) {
	v, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	return &appdto.JobView{
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
	}, nil
}
