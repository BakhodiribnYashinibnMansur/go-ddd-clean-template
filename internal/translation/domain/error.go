package domain

import shared "gct/internal/shared/domain"

var (
	ErrTranslationNotFound = shared.NewDomainError("TRANSLATION_NOT_FOUND", "translation not found")
)
