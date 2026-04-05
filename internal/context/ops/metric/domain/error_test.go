package domain_test

import (
	"testing"

	domain "gct/internal/context/ops/metric/domain"
)

func TestErrMetricNotFound(t *testing.T) {
	if domain.ErrMetricNotFound == nil {
		t.Fatal("ErrMetricNotFound should not be nil")
	}

	errMsg := domain.ErrMetricNotFound.Error()
	if errMsg == "" {
		t.Fatal("expected non-empty error message")
	}
}
