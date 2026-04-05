package domain_test

import (
	"testing"

	"gct/internal/kernel/domain"
)

func TestSortOrder_IsValid(t *testing.T) {
	tests := []struct {
		order domain.SortOrder
		valid bool
	}{
		{domain.SortOrderASC, true},
		{domain.SortOrderDESC, true},
		{domain.SortOrder("INVALID"), false},
		{domain.SortOrder(""), false},
		{domain.SortOrder("asc"), false},
	}

	for _, tt := range tests {
		if got := tt.order.IsValid(); got != tt.valid {
			t.Errorf("SortOrder(%q).IsValid() = %v, want %v", tt.order, got, tt.valid)
		}
	}
}

func TestPagination_GettersSetters(t *testing.T) {
	p := &domain.Pagination{}

	p.SetLimit(25)
	if p.GetLimit() != 25 {
		t.Errorf("expected limit 25, got %d", p.GetLimit())
	}

	p.SetOffset(50)
	if p.GetOffset() != 50 {
		t.Errorf("expected offset 50, got %d", p.GetOffset())
	}

	p.SetTotal(100)
	if p.GetTotal() != 100 {
		t.Errorf("expected total 100, got %d", p.GetTotal())
	}
}

func TestLang_GettersSetters(t *testing.T) {
	l := &domain.Lang{}

	l.SetUz("salom")
	l.SetRu("привет")
	l.SetEn("hello")

	if l.GetUz() != "salom" {
		t.Errorf("expected 'salom', got %q", l.GetUz())
	}
	if l.GetRu() != "привет" {
		t.Errorf("expected 'привет', got %q", l.GetRu())
	}
	if l.GetEn() != "hello" {
		t.Errorf("expected 'hello', got %q", l.GetEn())
	}
}

func TestFile_GettersSetters(t *testing.T) {
	f := &domain.File{}

	f.SetName("photo.jpg")
	f.SetLink("https://example.com/photo.jpg")

	if f.GetName() != "photo.jpg" {
		t.Errorf("expected 'photo.jpg', got %q", f.GetName())
	}
	if f.GetLink() != "https://example.com/photo.jpg" {
		t.Errorf("expected link, got %q", f.GetLink())
	}
}
