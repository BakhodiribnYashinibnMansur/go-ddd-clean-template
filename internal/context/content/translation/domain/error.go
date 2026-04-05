package domain

import shared "gct/internal/platform/domain"

// Sentinel domain errors for the Translation bounded context.
var (
	ErrTranslationNotFound = shared.NewDomainError("TRANSLATION_NOT_FOUND", "translation not found")
)
