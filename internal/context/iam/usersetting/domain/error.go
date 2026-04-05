package domain

import shared "gct/internal/kernel/domain"

// Sentinel domain errors for the UserSetting bounded context.
var (
	ErrUserSettingNotFound = shared.NewDomainError("USER_SETTING_NOT_FOUND", "user setting not found")
)
