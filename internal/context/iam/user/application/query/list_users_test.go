package query

import (
	"gct/internal/kernel/infrastructure/logger"
	"context"
	"testing"

	shared "gct/internal/kernel/domain"
	"gct/internal/context/iam/user/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListUsersHandler_Handle(t *testing.T) {
	t.Parallel()

	id1 := uuid.New()
	id2 := uuid.New()
	phone1 := "+998901111111"
	phone2 := "+998902222222"

	readRepo := &mockUserReadRepository{
		views: []*domain.UserView{
			{ID: id1, Phone: phone1, Active: true, IsApproved: true},
			{ID: id2, Phone: phone2, Active: true, IsApproved: false},
		},
		total: 2,
	}

	handler := NewListUsersHandler(readRepo, logger.Noop())

	result, err := handler.Handle(context.Background(), ListUsersQuery{
		Filter: domain.UsersFilter{},
	})
	require.NoError(t, err)

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}

	if len(result.Users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(result.Users))
	}

	if result.Users[0].ID != id1 {
		t.Errorf("expected first user ID %s, got %s", id1, result.Users[0].ID)
	}

	if result.Users[1].Phone != phone2 {
		t.Errorf("expected second user phone %s, got %s", phone2, result.Users[1].Phone)
	}
}

func TestListUsersHandler_Empty(t *testing.T) {
	t.Parallel()

	readRepo := &mockUserReadRepository{
		views: []*domain.UserView{},
		total: 0,
	}

	handler := NewListUsersHandler(readRepo, logger.Noop())

	result, err := handler.Handle(context.Background(), ListUsersQuery{
		Filter: domain.UsersFilter{},
	})
	require.NoError(t, err)

	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}

	if len(result.Users) != 0 {
		t.Errorf("expected 0 users, got %d", len(result.Users))
	}
}

func TestListUsersHandler_WithPagination(t *testing.T) {
	t.Parallel()

	readRepo := &mockUserReadRepository{
		views: []*domain.UserView{
			{ID: uuid.New(), Phone: "+998901111111", Active: true},
		},
		total: 5, // total is 5 but only 1 returned (limit=1)
	}

	handler := NewListUsersHandler(readRepo, logger.Noop())

	result, err := handler.Handle(context.Background(), ListUsersQuery{
		Filter: domain.UsersFilter{
			Pagination: &shared.Pagination{Limit: 1, Offset: 0},
		},
	})
	require.NoError(t, err)

	if result.Total != 5 {
		t.Errorf("expected total 5, got %d", result.Total)
	}
	if len(result.Users) != 1 {
		t.Errorf("expected 1 user in page, got %d", len(result.Users))
	}
}

func TestListUsersHandler_WithFilters(t *testing.T) {
	t.Parallel()

	activeUser := &domain.UserView{ID: uuid.New(), Phone: "+998901111111", Active: true, IsApproved: true}

	readRepo := &mockUserReadRepository{
		views: []*domain.UserView{activeUser},
		total: 1,
	}

	handler := NewListUsersHandler(readRepo, logger.Noop())

	active := true
	approved := true
	phone := "+998901111111"
	email := "test@example.com"

	result, err := handler.Handle(context.Background(), ListUsersQuery{
		Filter: domain.UsersFilter{
			Phone:      &phone,
			Email:      &email,
			Active:     &active,
			IsApproved: &approved,
			Pagination: &shared.Pagination{Limit: 10, Offset: 0},
		},
	})
	require.NoError(t, err)

	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
	if len(result.Users) != 1 {
		t.Errorf("expected 1 user, got %d", len(result.Users))
	}
	if result.Users[0].Active != true {
		t.Error("expected active user")
	}
}

func TestListUsersHandler_AllFieldsMapped(t *testing.T) {
	t.Parallel()

	roleID := uuid.New()
	email := "mapped@test.com"
	username := "mappeduser"

	readRepo := &mockUserReadRepository{
		views: []*domain.UserView{
			{
				ID:         uuid.New(),
				Phone:      "+998901234567",
				Email:      &email,
				Username:   &username,
				RoleID:     &roleID,
				Attributes: map[string]string{"key": "val"},
				Active:     true,
				IsApproved: true,
			},
		},
		total: 1,
	}

	handler := NewListUsersHandler(readRepo, logger.Noop())
	result, err := handler.Handle(context.Background(), ListUsersQuery{Filter: domain.UsersFilter{}})
	require.NoError(t, err)

	u := result.Users[0]
	if u.Email == nil || *u.Email != "mapped@test.com" {
		t.Error("email not mapped")
	}
	if u.Username == nil || *u.Username != "mappeduser" {
		t.Error("username not mapped")
	}
	if u.RoleID == nil || *u.RoleID != roleID {
		t.Error("roleID not mapped")
	}
	if u.Attributes["key"] != "val" {
		t.Error("attributes not mapped")
	}
}

func TestListUsersHandler_RepoError(t *testing.T) {
	t.Parallel()

	readRepo := &errorReadRepo{err: errRepoFailure}

	handler := NewListUsersHandler(readRepo, logger.Noop())
	_, err := handler.Handle(context.Background(), ListUsersQuery{Filter: domain.UsersFilter{}})
	if err == nil {
		t.Fatal("expected error from repo")
	}
}
