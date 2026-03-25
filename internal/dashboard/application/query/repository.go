package query

import (
	"context"

	appdto "gct/internal/dashboard/application"
)

// DashboardReadRepository defines the read-side persistence contract for dashboard stats.
type DashboardReadRepository interface {
	GetStats(ctx context.Context) (*appdto.DashboardStatsView, error)
}
