package query

import (
	"gct/internal/platform/infrastructure/logger"
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/context/iam/audit/domain"

	"github.com/google/uuid"
)

// --- Mock AuditReadRepository ---

type mockAuditReadRepository struct {
	auditLogs      []*domain.AuditLogView
	auditTotal     int64
	auditErr       error
	endpointViews  []*domain.EndpointHistoryView
	endpointTotal  int64
	endpointErr    error
}

func (m *mockAuditReadRepository) ListAuditLogs(_ context.Context, _ domain.AuditLogFilter) ([]*domain.AuditLogView, int64, error) {
	if m.auditErr != nil {
		return nil, 0, m.auditErr
	}
	return m.auditLogs, m.auditTotal, nil
}

func (m *mockAuditReadRepository) ListEndpointHistory(_ context.Context, _ domain.EndpointHistoryFilter) ([]*domain.EndpointHistoryView, int64, error) {
	if m.endpointErr != nil {
		return nil, 0, m.endpointErr
	}
	return m.endpointViews, m.endpointTotal, nil
}

// --- Tests ---

func TestListAuditLogsHandler_Success(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	readRepo := &mockAuditReadRepository{
		auditLogs: []*domain.AuditLogView{
			{
				ID:        uuid.New(),
				UserID:    &userID,
				Action:    domain.AuditActionLogin,
				Success:   true,
				CreatedAt: now,
			},
			{
				ID:        uuid.New(),
				UserID:    &userID,
				Action:    domain.AuditActionLogout,
				Success:   true,
				CreatedAt: now,
			},
		},
		auditTotal: 2,
	}

	handler := NewListAuditLogsHandler(readRepo, logger.Noop())

	q := ListAuditLogsQuery{
		Filter: domain.AuditLogFilter{},
	}

	result, err := handler.Handle(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}

	if len(result.AuditLogs) != 2 {
		t.Fatalf("expected 2 audit logs, got %d", len(result.AuditLogs))
	}

	if result.AuditLogs[0].Action != "LOGIN" {
		t.Errorf("expected action LOGIN, got %s", result.AuditLogs[0].Action)
	}

	if result.AuditLogs[1].Action != "LOGOUT" {
		t.Errorf("expected action LOGOUT, got %s", result.AuditLogs[1].Action)
	}

	if !result.AuditLogs[0].Success {
		t.Error("expected first log success to be true")
	}
}

func TestListAuditLogsHandler_Empty(t *testing.T) {
	readRepo := &mockAuditReadRepository{
		auditLogs:  []*domain.AuditLogView{},
		auditTotal: 0,
	}

	handler := NewListAuditLogsHandler(readRepo, logger.Noop())

	q := ListAuditLogsQuery{
		Filter: domain.AuditLogFilter{},
	}

	result, err := handler.Handle(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}

	if len(result.AuditLogs) != 0 {
		t.Errorf("expected 0 audit logs, got %d", len(result.AuditLogs))
	}
}

func TestListAuditLogsHandler_RepoError(t *testing.T) {
	readRepo := &mockAuditReadRepository{
		auditErr: errors.New("database unavailable"),
	}

	handler := NewListAuditLogsHandler(readRepo, logger.Noop())

	q := ListAuditLogsQuery{
		Filter: domain.AuditLogFilter{},
	}

	result, err := handler.Handle(context.Background(), q)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestListAuditLogsHandler_MapsAllFields(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	resourceID := uuid.New()
	policyID := uuid.New()
	resType := "user"
	platform := "web"
	ip := "10.0.0.1"
	ua := "TestAgent"
	perm := "users:write"
	decision := "deny"
	errMsg := "access denied"
	now := time.Now()

	readRepo := &mockAuditReadRepository{
		auditLogs: []*domain.AuditLogView{
			{
				ID:           uuid.New(),
				UserID:       &userID,
				SessionID:    &sessionID,
				Action:       domain.AuditActionAccessDenied,
				ResourceType: &resType,
				ResourceID:   &resourceID,
				Platform:     &platform,
				IPAddress:    &ip,
				UserAgent:    &ua,
				Permission:   &perm,
				PolicyID:     &policyID,
				Decision:     &decision,
				Success:      false,
				ErrorMessage: &errMsg,
				Metadata:     map[string]string{"reason": "role_mismatch"},
				CreatedAt:    now,
			},
		},
		auditTotal: 1,
	}

	handler := NewListAuditLogsHandler(readRepo, logger.Noop())

	result, err := handler.Handle(context.Background(), ListAuditLogsQuery{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	v := result.AuditLogs[0]

	if *v.UserID != userID {
		t.Error("userID mismatch")
	}
	if *v.SessionID != sessionID {
		t.Error("sessionID mismatch")
	}
	if v.Action != "ACCESS_DENIED" {
		t.Errorf("expected ACCESS_DENIED, got %s", v.Action)
	}
	if *v.ResourceType != "user" {
		t.Error("resourceType mismatch")
	}
	if *v.ResourceID != resourceID {
		t.Error("resourceID mismatch")
	}
	if *v.Platform != "web" {
		t.Error("platform mismatch")
	}
	if *v.IPAddress != "10.0.0.1" {
		t.Error("ipAddress mismatch")
	}
	if *v.UserAgent != "TestAgent" {
		t.Error("userAgent mismatch")
	}
	if *v.Permission != "users:write" {
		t.Error("permission mismatch")
	}
	if *v.PolicyID != policyID {
		t.Error("policyID mismatch")
	}
	if *v.Decision != "deny" {
		t.Error("decision mismatch")
	}
	if v.Success {
		t.Error("expected success to be false")
	}
	if *v.ErrorMessage != "access denied" {
		t.Error("errorMessage mismatch")
	}
	if v.Metadata["reason"] != "role_mismatch" {
		t.Error("metadata mismatch")
	}
	if v.CreatedAt.IsZero() {
		t.Error("createdAt should not be zero")
	}
}

func TestListAuditLogsHandler_WithFilter(t *testing.T) {
	userID := uuid.New()
	action := domain.AuditActionLogin
	success := true

	readRepo := &mockAuditReadRepository{
		auditLogs: []*domain.AuditLogView{
			{
				ID:      uuid.New(),
				UserID:  &userID,
				Action:  domain.AuditActionLogin,
				Success: true,
			},
		},
		auditTotal: 1,
	}

	handler := NewListAuditLogsHandler(readRepo, logger.Noop())

	q := ListAuditLogsQuery{
		Filter: domain.AuditLogFilter{
			UserID:  &userID,
			Action:  &action,
			Success: &success,
		},
	}

	result, err := handler.Handle(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}
