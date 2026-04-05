package domain

import shared "gct/internal/kernel/domain"

// Domain errors for the iprule bounded context.
// Returned by repositories when the requested IP rule does not exist in the data store.
var (
	ErrIPRuleNotFound = shared.NewDomainError("IP_RULE_NOT_FOUND", "ip rule not found")
)
