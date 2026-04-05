package validation

// IsEnumValid checks if a value exists in the allowed list of values.
func IsEnumValid[T comparable](val T, allowed []T) bool {
	for _, a := range allowed {
		if val == a {
			return true
		}
	}
	return false
}
