package domain

const (
	OpEq       = "eq"
	OpNotEq    = "not_eq"
	OpIn       = "in"
	OpNotIn    = "not_in"
	OpGt       = "gt"
	OpGte      = "gte"
	OpLt       = "lt"
	OpLte      = "lte"
	OpContains = "contains"
)

var ValidOperators = map[string]bool{
	OpEq: true, OpNotEq: true,
	OpIn: true, OpNotIn: true,
	OpGt: true, OpGte: true,
	OpLt: true, OpLte: true,
	OpContains: true,
}

func IsValidOperator(op string) bool {
	return ValidOperators[op]
}
