package entity

import shared "gct/internal/kernel/domain"

// Sentinel domain errors for the SiteSetting bounded context.
// Use errors.Is to match these in the application/presentation layers.
var (
	ErrSiteSettingNotFound = shared.NewDomainError("SITE_SETTING_NOT_FOUND", "site setting not found")
)
