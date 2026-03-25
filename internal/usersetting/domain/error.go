package domain

import shared "gct/internal/shared/domain"

var (
	ErrUserSettingNotFound = shared.NewDomainError("USER_SETTING_NOT_FOUND", "user setting not found")
)
