package entity

import shared "gct/internal/kernel/domain"

// Domain errors for the dataexport bounded context.
var (
	// ErrDataExportNotFound signals that no export record exists for the requested identifier.
	// Repository implementations must return this sentinel so the application layer can map it to HTTP 404.
	ErrDataExportNotFound = shared.NewDomainError("DATA_EXPORT_NOT_FOUND", "data export not found")
)
