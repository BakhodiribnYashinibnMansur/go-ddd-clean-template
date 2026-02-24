package job

import (
	"context"

	"gct/internal/domain"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, j *domain.Job) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error)
	List(ctx context.Context, filter domain.JobFilter) ([]domain.Job, int64, error)
	Update(ctx context.Context, j *domain.Job) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UseCaseI interface {
	Create(ctx context.Context, req domain.CreateJobRequest) (*domain.Job, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error)
	List(ctx context.Context, filter domain.JobFilter) ([]domain.Job, int64, error)
	Update(ctx context.Context, id uuid.UUID, req domain.UpdateJobRequest) (*domain.Job, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Trigger(ctx context.Context, id uuid.UUID) (*domain.Job, error)
}
