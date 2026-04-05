package domain

import shared "gct/internal/kernel/domain"

// Domain errors for the metric bounded context.
// Primarily used when querying individual metric records by ID.
var (
	ErrMetricNotFound = shared.NewDomainError("METRIC_NOT_FOUND", "metric not found")
)
