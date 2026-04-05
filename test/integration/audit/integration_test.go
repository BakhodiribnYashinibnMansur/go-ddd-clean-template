package audit

import (
	"context"
	"testing"

	"gct/internal/context/iam/supporting/audit"
	"gct/internal/context/iam/supporting/audit/application/command"
	"gct/internal/context/iam/supporting/audit/application/query"
	"gct/internal/context/iam/supporting/audit/domain"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/eventbus"
	"gct/internal/kernel/infrastructure/logger"
	"gct/test/integration/common/setup"
)

func newTestBC(t *testing.T) *audit.BoundedContext {
	t.Helper()
	eb := eventbus.NewInMemoryEventBus()
	l := logger.New("error")
	return audit.NewBoundedContext(setup.TestPG.Pool, eb, l)
}

func TestIntegration_CreateAuditLogAndList(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	ip := "192.168.1.1"
	ua := "test-agent/1.0"
	err := bc.CreateAuditLog.Handle(ctx, command.CreateAuditLogCommand{
		Action:    domain.AuditActionLogin,
		IPAddress: &ip,
		UserAgent: &ua,
		Success:   true,
		Metadata:  map[string]string{"source": "integration-test"},
	})
	if err != nil {
		t.Fatalf("CreateAuditLog: %v", err)
	}

	result, err := bc.ListAuditLogs.Handle(ctx, query.ListAuditLogsQuery{
		Filter: domain.AuditLogFilter{
			Pagination: &shared.Pagination{Limit: 10, Offset: 0},
		},
	})
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 audit log, got %d", result.Total)
	}

	a := result.AuditLogs[0]
	if a.Action != string(domain.AuditActionLogin) {
		t.Errorf("expected action LOGIN, got %s", a.Action)
	}
	if !a.Success {
		t.Error("expected success to be true")
	}
}

func TestIntegration_CreateEndpointHistoryAndList(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	ip := "10.0.0.1"
	ua := "curl/7.68.0"
	err := bc.CreateEndpointHistory.Handle(ctx, command.CreateEndpointHistoryCommand{
		Endpoint:   "/api/v1/users",
		Method:     "GET",
		StatusCode: 200,
		Latency:    42,
		IPAddress:  &ip,
		UserAgent:  &ua,
	})
	if err != nil {
		t.Fatalf("CreateEndpointHistory: %v", err)
	}

	result, err := bc.ListEndpointHistory.Handle(ctx, query.ListEndpointHistoryQuery{
		Filter: domain.EndpointHistoryFilter{
			Pagination: &shared.Pagination{Limit: 10, Offset: 0},
		},
	})
	if err != nil {
		t.Fatalf("ListEndpointHistory: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected 1 endpoint history entry, got %d", result.Total)
	}

	e := result.Entries[0]
	if e.Endpoint != "/api/v1/users" {
		t.Errorf("expected endpoint /api/v1/users, got %s", e.Endpoint)
	}
	if e.Method != "GET" {
		t.Errorf("expected method GET, got %s", e.Method)
	}
	if e.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", e.StatusCode)
	}
	if e.Latency != 42 {
		t.Errorf("expected latency 42, got %d", e.Latency)
	}
}
