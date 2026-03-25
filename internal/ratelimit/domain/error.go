package domain

import shared "gct/internal/shared/domain"

var (
	ErrRateLimitNotFound = shared.NewDomainError("RATE_LIMIT_NOT_FOUND", "rate limit not found")
)
