package domain

import shared "gct/internal/shared/domain"

var (
	ErrFileNotFound = shared.NewDomainError("FILE_NOT_FOUND", "file not found")
)
