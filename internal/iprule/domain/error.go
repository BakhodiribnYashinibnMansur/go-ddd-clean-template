package domain

import shared "gct/internal/shared/domain"

var (
	ErrIPRuleNotFound = shared.NewDomainError("IP_RULE_NOT_FOUND", "ip rule not found")
)
