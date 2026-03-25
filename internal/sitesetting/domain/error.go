package domain

import shared "gct/internal/shared/domain"

var (
	ErrSiteSettingNotFound = shared.NewDomainError("SITE_SETTING_NOT_FOUND", "site setting not found")
)
