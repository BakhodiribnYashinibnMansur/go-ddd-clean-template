package domain

import shared "gct/internal/shared/domain"

var (
	ErrErrorCodeNotFound = shared.NewDomainError("ERROR_CODE_NOT_FOUND", "error code not found")
)
