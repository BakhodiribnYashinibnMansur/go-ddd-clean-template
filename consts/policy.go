package consts

// Policy Condition Keys
const (
	PolicyKeyIP        string = "ip"
	PolicyKeyUserAgent string = "user_agent"
	PolicyKeyTime      string = "time"
	PolicyKeyUserID    string = "user_id"
	PolicyKeyRoleID    string = "role_id"

	// Operators
	OpEquals          string = "equals"
	OpNotEquals       string = "not_equals"
	OpIn              string = "in"
	OpNotIn           string = "not_in"
	OpContains        string = "contains"
	OpAny             string = "any"
	OpAll             string = "all"
	OpBetween         string = "between"
	OpGreaterThan     string = "gt"
	OpLessThan        string = "lt"
	OpGreaterOrEquals string = "gte"
	OpLessOrEquals    string = "lte"

	// Logical Operators
	OpAnd string = "and"
	OpOr  string = "or"
	OpNot string = "not"

	// Special Keys & Namespaces
	KeyTarget   string = "target"
	KeyUser     string = "user"
	KeyEnv      string = "env"
	KeyResource string = "resource"
	KeyRelation string = "relation"
)

// List of allowed keys for validation
var AllowedPolicyKeys = map[string]bool{
	// Legacy/Direct keys
	PolicyKeyIP:        true,
	PolicyKeyUserAgent: true,
	PolicyKeyTime:      true,
	PolicyKeyUserID:    true,
	PolicyKeyRoleID:    true,
	KeyTarget:          true,

	// Namespaces
	KeyUser:     true,
	KeyEnv:      true,
	KeyResource: true,
	KeyRelation: true,

	// Logical
	OpAnd: true,
	OpOr:  true,
	OpNot: true,
	// dynamic keys from router paths
	ParamID:         true,
	ParamSessionID:  true,
	ParamPermID:     true,
	ParamPolicyID:   true,
	ParamRelationID: true,
}
