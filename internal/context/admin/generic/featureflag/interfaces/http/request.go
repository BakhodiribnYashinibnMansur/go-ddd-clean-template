package http

// CreateRequest represents the request body for creating a feature flag.
type CreateRequest struct {
	Name              string `json:"name" binding:"required,max=255"`
	Key               string `json:"key" binding:"required,max=128"`
	Description       string `json:"description" binding:"max=2000"`
	FlagType          string `json:"flag_type" binding:"required,max=50"`
	DefaultValue      string `json:"default_value" binding:"max=1000"`
	RolloutPercentage int    `json:"rollout_percentage" binding:"min=0,max=100"`
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
	Name       string             `json:"name" binding:"required,max=255"`
	Variation  string             `json:"variation" binding:"required,max=1000"`
	Priority   int                `json:"priority"`
	Conditions []ConditionRequest `json:"conditions" binding:"max=50"`
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
	Attribute string `json:"attribute" binding:"required,max=255"`
	Operator  string `json:"operator" binding:"required,max=50"`
	Value     string `json:"value" binding:"required,max=1000"`
}

// EvaluateRequest represents the request body for evaluating a single feature flag.
type EvaluateRequest struct {
	Key       string            `json:"key" binding:"required,max=128"`
	UserAttrs map[string]string `json:"user_attrs"`
}

// BatchEvaluateRequest represents the request body for evaluating multiple feature flags.
type BatchEvaluateRequest struct {
	Keys      []string          `json:"keys" binding:"required,min=1,max=100"`
	UserAttrs map[string]string `json:"user_attrs"`
}
