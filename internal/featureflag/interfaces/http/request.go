package http

// CreateRequest represents the request body for creating a feature flag.
type CreateRequest struct {
	Name              string `json:"name" binding:"required"`
	Key               string `json:"key" binding:"required"`
	Description       string `json:"description"`
	FlagType          string `json:"flag_type" binding:"required"`
	DefaultValue      string `json:"default_value"`
	RolloutPercentage int    `json:"rollout_percentage"`
	IsActive          bool   `json:"is_active"`
}

// UpdateRequest represents the request body for updating a feature flag.
type UpdateRequest struct {
	Name              *string `json:"name,omitempty"`
	Key               *string `json:"key,omitempty"`
	Description       *string `json:"description,omitempty"`
	FlagType          *string `json:"flag_type,omitempty"`
	DefaultValue      *string `json:"default_value,omitempty"`
	RolloutPercentage *int    `json:"rollout_percentage,omitempty"`
	IsActive          *bool   `json:"is_active,omitempty"`
}

// CreateRuleGroupRequest represents the request body for creating a rule group.
type CreateRuleGroupRequest struct {
	Name       string             `json:"name" binding:"required"`
	Variation  string             `json:"variation" binding:"required"`
	Priority   int                `json:"priority"`
	Conditions []ConditionRequest `json:"conditions"`
}

// UpdateRuleGroupRequest represents the request body for updating a rule group.
type UpdateRuleGroupRequest struct {
	Name       *string             `json:"name,omitempty"`
	Variation  *string             `json:"variation,omitempty"`
	Priority   *int                `json:"priority,omitempty"`
	Conditions *[]ConditionRequest `json:"conditions,omitempty"`
}

// ConditionRequest represents a single targeting condition in a rule group.
type ConditionRequest struct {
	Attribute string `json:"attribute" binding:"required"`
	Operator  string `json:"operator" binding:"required"`
	Value     string `json:"value" binding:"required"`
}
