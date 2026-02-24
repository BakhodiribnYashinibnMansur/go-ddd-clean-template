package dashboard

import (
	"context"

	"gct/internal/domain"
)

// Repository defines the data access methods needed by the dashboard use case.
type Repository interface {
	Get(ctx context.Context) (domain.DashboardStats, error)
}

// UseCaseI defines the business logic interface for dashboard stats.
type UseCaseI interface {
	Get(ctx context.Context) (domain.DashboardStats, error)
}
