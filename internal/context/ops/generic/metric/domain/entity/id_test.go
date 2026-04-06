package entity_test

import (
	"testing"

	"gct/internal/context/ops/generic/metric/domain/entity"

	"github.com/google/uuid"
)

func TestMetricID_RoundTrip(t *testing.T) {
	t.Parallel()

	id := entity.NewMetricID()
	if id.IsZero() {
		t.Fatal("newly generated MetricID should not be zero")
	}

	parsed, err := entity.ParseMetricID(id.String())
	if err != nil {
		t.Fatalf("ParseMetricID round-trip failed: %v", err)
	}
	if parsed != id {
		t.Fatalf("round-trip mismatch: got %s, want %s", parsed, id)
	}
	if parsed.UUID() != id.UUID() {
		t.Fatalf("UUID() mismatch")
	}
}

func TestParseMetricID_Invalid(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, in string }{
		{"empty", ""},
		{"garbage", "not-a-uuid"},
		{"truncated", "123e4567-e89b-12d3-a456"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := entity.ParseMetricID(tc.in); err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
		})
	}
}

func TestMetricID_IsZero(t *testing.T) {
	t.Parallel()

	var zero entity.MetricID
	if !zero.IsZero() {
		t.Fatal("zero-valued MetricID should report IsZero()")
	}

	nonZero := entity.MetricID(uuid.New())
	if nonZero.IsZero() {
		t.Fatal("non-zero MetricID should not report IsZero()")
	}
}

func TestMetricID_Distinct(t *testing.T) {
	t.Parallel()

	a := entity.NewMetricID()
	b := entity.NewMetricID()
	if a == b {
		t.Fatal("separately generated IDs should differ")
	}
}
