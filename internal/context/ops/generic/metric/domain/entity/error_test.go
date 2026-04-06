package entity_test

import (
	"testing"

	"gct/internal/context/ops/generic/metric/domain/entity"
)

func TestErrMetricNotFound(t *testing.T) {
	if entity.ErrMetricNotFound == nil {
		t.Fatal("ErrMetricNotFound should not be nil")
	}

	errMsg := entity.ErrMetricNotFound.Error()
	if errMsg == "" {
		t.Fatal("expected non-empty error message")
	}
}
