package domain

import shared "gct/internal/shared/domain"

var (
	ErrMetricNotFound = shared.NewDomainError("METRIC_NOT_FOUND", "metric not found")
)
