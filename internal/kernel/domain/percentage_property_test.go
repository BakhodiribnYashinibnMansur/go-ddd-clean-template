package domain_test

import (
	"fmt"
	"testing"

	"gct/internal/kernel/domain"

	"pgregory.net/rapid"
)

func TestPercentage_Property_RangeAcceptance(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		v := rapid.IntRange(0, 100).Draw(t, "value")
		_, err := domain.NewPercentage(v)
		if err != nil {
			t.Fatalf("valid value %d rejected: %v", v, err)
		}
	})
}

func TestPercentage_Property_RangeRejection(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		v := rapid.OneOf(
			rapid.IntRange(-1_000_000, -1),
			rapid.IntRange(101, 1_000_000),
		).Draw(t, "invalid_value")
		_, err := domain.NewPercentage(v)
		if err == nil {
			t.Fatalf("invalid value %d accepted", v)
		}
	})
}

func TestPercentage_Property_Roundtrip(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		v := rapid.IntRange(0, 100).Draw(t, "value")
		p, err := domain.NewPercentage(v)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Int() != v {
			t.Fatalf("Int() = %d, want %d", p.Int(), v)
		}
	})
}

func TestPercentage_Property_StringFormat(t *testing.T) {
	t.Parallel()
	rapid.Check(t, func(t *rapid.T) {
		v := rapid.IntRange(0, 100).Draw(t, "value")
		p, err := domain.NewPercentage(v)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := fmt.Sprintf("%d%%", v)
		if p.String() != want {
			t.Fatalf("String() = %q, want %q", p.String(), want)
		}
	})
}
