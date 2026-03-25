package domain

import shared "gct/internal/shared/domain"

var (
	ErrSystemErrorNotFound = shared.NewDomainError("SYSTEM_ERROR_NOT_FOUND", "system error not found")
)
