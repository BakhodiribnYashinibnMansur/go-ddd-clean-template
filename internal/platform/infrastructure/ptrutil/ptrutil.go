package ptrutil

// StrVal returns the string value from a pointer, or empty string if nil.
func StrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// IntVal returns the int value from a pointer, or 0 if nil.
func IntVal(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// BoolVal returns the bool value from a pointer, or false if nil.
func BoolVal(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// Ptr returns a pointer to the given value.
func Ptr[T any](v T) *T {
	return &v
}
