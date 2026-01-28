package mock

import (
	"time"

	"gct/internal/domain"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User generates a fake domain.User
func User() *domain.User {
	user := &domain.User{
		ID:        uuid.New(),
		Username:  func() *string { s := gofakeit.Name(); return &s }(),
		Phone:     func() *string { phone := gofakeit.Phone(); return &phone }(),
		Salt:      func() *string { s := gofakeit.LetterN(10); return &s }(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: 0,
		LastSeen:  func() *time.Time { t := time.Now(); return &t }(),
	}

	// Set password hash
	password := gofakeit.Password(true, true, true, true, false, 12)
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user.PasswordHash = string(hash)

	return user
}

// Users generates multiple fake domain.User
func Users(count int) []*domain.User {
	users := make([]*domain.User, count)
	for i := range count {
		users[i] = User()
	}
	return users
}

// UserWithPassword generates a fake domain.User with password set
func UserWithPassword() *domain.User {
	user := User()
	password := gofakeit.Password(true, true, true, true, false, 12)
	_ = user.SetPassword(password) // Mock data, ignore error
	return user
}

// UserFilter generates a fake domain.UserFilter
func UserFilter() *domain.UserFilter {
	return &domain.UserFilter{
		ID:    func() *uuid.UUID { id := uuid.New(); return &id }(),
		Phone: func() *string { phone := gofakeit.Phone(); return &phone }(),
	}
}

// UserFilterWithID generates a fake domain.UserFilter with ID
func UserFilterWithID(id uuid.UUID) *domain.UserFilter {
	return &domain.UserFilter{
		ID:    &id,
		Phone: func() *string { phone := gofakeit.Phone(); return &phone }(),
	}
}

// UserFilterWithPhone generates a fake domain.UserFilter with Phone
func UserFilterWithPhone(phone string) *domain.UserFilter {
	return &domain.UserFilter{
		ID:    func() *uuid.UUID { id := uuid.New(); return &id }(),
		Phone: &phone,
	}
}

// UsersFilter generates a fake domain.UsersFilter
func UsersFilter() *domain.UsersFilter {
	return &domain.UsersFilter{
		UserFilter: *UserFilter(),
		Pagination: Pagination(),
	}
}

// UsersFilterWithPagination generates a fake domain.UsersFilter with custom pagination
func UsersFilterWithPagination(limit, offset, total int64) *domain.UsersFilter {
	return &domain.UsersFilter{
		UserFilter: *UserFilter(),
		Pagination: PaginationWithValues(limit, offset, total),
	}
}
