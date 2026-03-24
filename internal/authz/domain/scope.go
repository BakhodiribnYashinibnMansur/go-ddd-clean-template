package domain

// Scope is a value object representing an API endpoint.
type Scope struct {
	Path   string
	Method string // GET, POST, PUT, DELETE, PATCH
}
