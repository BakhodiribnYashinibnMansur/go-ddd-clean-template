package domain

import shared "gct/internal/shared/domain"

// Domain errors for the job bounded context.
// Returned by repositories when the requested job does not exist in the data store.
var (
	ErrJobNotFound = shared.NewDomainError("JOB_NOT_FOUND", "job not found")
)
