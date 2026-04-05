package domain

import shared "gct/internal/kernel/domain"

var (
	ErrFeatureFlagNotFound = shared.NewDomainError("FEATURE_FLAG_NOT_FOUND", "feature flag not found")
	ErrRuleGroupNotFound   = shared.NewDomainError("RULE_GROUP_NOT_FOUND", "rule group not found")
	ErrInvalidOperator     = shared.NewDomainError("INVALID_OPERATOR", "invalid condition operator")
	ErrDuplicateKey        = shared.NewDomainError("DUPLICATE_FLAG_KEY", "feature flag key already exists")
)
