package domain

import shared "gct/internal/shared/domain"

var (
	ErrDataExportNotFound = shared.NewDomainError("DATA_EXPORT_NOT_FOUND", "data export not found")
)
