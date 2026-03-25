package domain

import shared "gct/internal/shared/domain"

var (
	ErrAnnouncementNotFound = shared.NewDomainError("ANNOUNCEMENT_NOT_FOUND", "announcement not found")
	ErrAlreadyPublished     = shared.NewDomainError("ANNOUNCEMENT_ALREADY_PUBLISHED", "announcement is already published")
)
