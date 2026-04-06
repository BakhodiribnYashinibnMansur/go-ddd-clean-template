package entity

import shared "gct/internal/kernel/domain"

// Domain errors for the ratelimit bounded context.
// Returned by repositories when the requested rate limit rule does not exist in the data store.
var (
	ErrRateLimitNotFound = shared.NewDomainError("RATE_LIMIT_NOT_FOUND", "rate limit not found")
)
