package domain

import shared "gct/internal/shared/domain"

var (
	ErrJobNotFound = shared.NewDomainError("JOB_NOT_FOUND", "job not found")
)
