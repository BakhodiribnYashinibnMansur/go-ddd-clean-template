package domain

// Scope is an immutable value object representing a single API endpoint identified by path and HTTP method.
// Two scopes are equal when both Path and Method match. Scope has no identity of its own —
// it derives meaning only as part of a Permission's scope list.
type Scope struct {
	Path   string
	Method string // GET, POST, PUT, DELETE, PATCH
}
