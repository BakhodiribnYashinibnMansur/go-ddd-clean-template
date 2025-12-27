package client

// UserFilter represents filters for retrieving users.
type UserFilter struct {
	ID    *int64
	Phone *string
}

// UserListFilter represents filters and pagination for retrieving multiple users.
type UserListFilter struct {
	UserFilter
	Limit  int
	Offset int
}
