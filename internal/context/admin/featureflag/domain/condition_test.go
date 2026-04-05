package domain_test

import (
	"testing"

	"gct/internal/context/admin/featureflag/domain"
)

func TestCondition_Match_Eq(t *testing.T) {
	t.Parallel()

	c := domain.NewCondition("country", "eq", "US")
	if !c.Match("US") {
		t.Fatal("expected match")
	}
	if c.Match("UK") {
		t.Fatal("expected no match")
	}
}

func TestCondition_Match_NotEq(t *testing.T) {
	t.Parallel()

	c := domain.NewCondition("country", "not_eq", "US")
	if !c.Match("UK") {
		t.Fatal("expected match")
	}
	if c.Match("US") {
		t.Fatal("expected no match")
	}
}

func TestCondition_Match_In(t *testing.T) {
	t.Parallel()

	c := domain.NewCondition("country", "in", "US,UK,CA")
	if !c.Match("UK") {
		t.Fatal("expected match")
	}
	if c.Match("DE") {
		t.Fatal("expected no match")
	}
}

func TestCondition_Match_NotIn(t *testing.T) {
	t.Parallel()

	c := domain.NewCondition("country", "not_in", "US,UK,CA")
	if !c.Match("DE") {
		t.Fatal("expected match")
	}
	if c.Match("US") {
		t.Fatal("expected no match")
	}
}

func TestCondition_Match_Gt(t *testing.T) {
	t.Parallel()

	c := domain.NewCondition("age", "gt", "18")
	if !c.Match("19") {
		t.Fatal("expected match")
	}
	if c.Match("18") {
		t.Fatal("expected no match")
	}
}

func TestCondition_Match_Gte(t *testing.T) {
	t.Parallel()

	c := domain.NewCondition("age", "gte", "18")
	if !c.Match("18") {
		t.Fatal("expected match")
	}
	if c.Match("17") {
		t.Fatal("expected no match")
	}
}

func TestCondition_Match_Lt(t *testing.T) {
	t.Parallel()

	c := domain.NewCondition("age", "lt", "18")
	if !c.Match("17") {
		t.Fatal("expected match")
	}
	if c.Match("18") {
		t.Fatal("expected no match")
	}
}

func TestCondition_Match_Lte(t *testing.T) {
	t.Parallel()

	c := domain.NewCondition("age", "lte", "18")
	if !c.Match("18") {
		t.Fatal("expected match")
	}
	if c.Match("19") {
		t.Fatal("expected no match")
	}
}

func TestCondition_Match_Contains(t *testing.T) {
	t.Parallel()

	c := domain.NewCondition("email", "contains", "@example.com")
	if !c.Match("user@example.com") {
		t.Fatal("expected match")
	}
	if c.Match("user@other.com") {
		t.Fatal("expected no match")
	}
}

func TestCondition_Match_InvalidOperator(t *testing.T) {
	t.Parallel()

	c := domain.NewCondition("country", "invalid", "US")
	if c.Match("US") {
		t.Fatal("expected no match for invalid operator")
	}
}

func TestCondition_Match_NonNumeric_Gt(t *testing.T) {
	t.Parallel()

	c := domain.NewCondition("age", "gt", "18")
	if c.Match("abc") {
		t.Fatal("expected no match for non-numeric value")
	}
}
