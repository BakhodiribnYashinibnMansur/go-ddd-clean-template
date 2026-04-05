package infrastructure_test

import (
	"testing"

	"gct/internal/kernel/infrastructure"
)

func TestBaseRepository_TableName(t *testing.T) {
	repo := infrastructure.NewBaseRepository[any](nil, "users", []string{"id", "phone"}, nil)
	if repo.TableName() != "users" {
		t.Errorf("got %s, want users", repo.TableName())
	}
}

func TestBaseRepository_Columns(t *testing.T) {
	repo := infrastructure.NewBaseRepository[any](nil, "users", []string{"id", "phone", "email"}, nil)
	if len(repo.Columns()) != 3 {
		t.Errorf("got %d, want 3", len(repo.Columns()))
	}
}

func TestBaseRepository_Builder(t *testing.T) {
	repo := infrastructure.NewBaseRepository[any](nil, "users", []string{"id"}, nil)
	// Builder should use Dollar placeholder format
	sql, _, err := repo.Builder().Select("*").From("test").Where("id = ?", 1).ToSql()
	if err != nil {
		t.Fatal(err)
	}
	if sql != "SELECT * FROM test WHERE id = $1" {
		t.Errorf("got %s, want Dollar format", sql)
	}
}
