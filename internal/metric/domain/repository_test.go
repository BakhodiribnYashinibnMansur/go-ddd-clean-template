package domain_test

import (
	"testing"
	"time"

	domain "gct/internal/metric/domain"

	"github.com/google/uuid"
)

func TestMetricFilter_ZeroValue(t *testing.T) {
	f := domain.MetricFilter{}

	if f.Name != nil {
		t.Fatal("expected nil Name")
	}
	if f.IsPanic != nil {
		t.Fatal("expected nil IsPanic")
	}
	if f.FromDate != nil {
		t.Fatal("expected nil FromDate")
	}
	if f.ToDate != nil {
		t.Fatal("expected nil ToDate")
	}
	if f.Limit != 0 {
		t.Fatalf("expected 0 Limit, got %d", f.Limit)
	}
	if f.Offset != 0 {
		t.Fatalf("expected 0 Offset, got %d", f.Offset)
	}
}

func TestMetricFilter_WithValues(t *testing.T) {
	name := "UserService.Create"
	isPanic := true
	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()

	f := domain.MetricFilter{
		Name:     &name,
		IsPanic:  &isPanic,
		FromDate: &from,
		ToDate:   &to,
		Limit:    50,
		Offset:   10,
	}

	if *f.Name != name {
		t.Fatalf("expected name %s, got %s", name, *f.Name)
	}
	if *f.IsPanic != isPanic {
		t.Fatalf("expected isPanic %v, got %v", isPanic, *f.IsPanic)
	}
	if f.Limit != 50 {
		t.Fatalf("expected limit 50, got %d", f.Limit)
	}
	if f.Offset != 10 {
		t.Fatalf("expected offset 10, got %d", f.Offset)
	}
}

func TestMetricView_Fields(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	panicErr := "nil pointer dereference"

	view := domain.MetricView{
		ID:         id,
		Name:       "AuthService.Login",
		LatencyMs:  42.5,
		IsPanic:    true,
		PanicError: &panicErr,
		CreatedAt:  now,
	}

	if view.ID != id {
		t.Fatal("ID mismatch")
	}
	if view.Name != "AuthService.Login" {
		t.Fatalf("expected AuthService.Login, got %s", view.Name)
	}
	if view.LatencyMs != 42.5 {
		t.Fatalf("expected 42.5, got %f", view.LatencyMs)
	}
	if !view.IsPanic {
		t.Fatal("expected isPanic true")
	}
	if view.PanicError == nil || *view.PanicError != panicErr {
		t.Fatal("panicError mismatch")
	}
	if view.CreatedAt != now {
		t.Fatal("createdAt mismatch")
	}
}
