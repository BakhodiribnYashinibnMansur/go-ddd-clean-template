package statistics

import (
	"testing"

	"gct/internal/kernel/infrastructure/logger"
)

func TestNewBoundedContext(t *testing.T) {
	bc := NewBoundedContext(nil, logger.Noop())
	if bc == nil {
		t.Fatal("expected non-nil BoundedContext")
	}
	if bc.GetOverview == nil {
		t.Error("GetOverview handler not wired")
	}
	if bc.GetUserStats == nil {
		t.Error("GetUserStats handler not wired")
	}
	if bc.GetSessionStats == nil {
		t.Error("GetSessionStats handler not wired")
	}
	if bc.GetErrorStats == nil {
		t.Error("GetErrorStats handler not wired")
	}
	if bc.GetAuditStats == nil {
		t.Error("GetAuditStats handler not wired")
	}
	if bc.GetSecurityStats == nil {
		t.Error("GetSecurityStats handler not wired")
	}
	if bc.GetFeatureFlagStats == nil {
		t.Error("GetFeatureFlagStats handler not wired")
	}
	if bc.GetContentStats == nil {
		t.Error("GetContentStats handler not wired")
	}
	if bc.GetIntegrationStats == nil {
		t.Error("GetIntegrationStats handler not wired")
	}
}
