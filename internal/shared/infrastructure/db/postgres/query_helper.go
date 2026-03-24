package postgres

import (
	"strconv"
	"strings"
)

// ParseNamedQuery converts named parameters in SQL query to numbered parameters.
// Example: "SELECT * FROM users WHERE id = :id AND name = :name"
// becomes: "SELECT * FROM users WHERE id = $1 AND name = $2"
// and returns the corresponding parameter values in order.
func ParseNamedQuery(query *string, args map[string]any) (params []any) {
	count := 0
	params = []any{}
	for key, value := range args {
		if strings.Contains(*query, key) {
			params = append(params, value)
			count++
			*query = strings.ReplaceAll(*query, key, " $"+strconv.Itoa(count)+" ")
		}
	}
	return params
}
