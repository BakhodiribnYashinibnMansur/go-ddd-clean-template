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

func TestListEndpointHistoryHandler_Success(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	ip := "192.168.1.1"
	ua := "Mozilla/5.0"

	readRepo := &mockAuditReadRepository{
		endpointViews: []*domain.EndpointHistoryView{
			{
				ID:         uuid.New(),
				UserID:     &userID,
				Endpoint:   "/api/v1/users",
				Method:     "GET",
				StatusCode: 200,
				Latency:    42,
				IPAddress:  &ip,
				UserAgent:  &ua,
				CreatedAt:  now,
			},
			{
				ID:         uuid.New(),
				UserID:     &userID,
				Endpoint:   "/api/v1/roles",
				Method:     "POST",
				StatusCode: 201,
				Latency:    88,
				IPAddress:  &ip,
				UserAgent:  &ua,
				CreatedAt:  now,
			},
		},
		endpointTotal: 2,
	}

	handler := NewListEndpointHistoryHandler(readRepo, logger.Noop())

	q := ListEndpointHistoryQuery{
		Filter: domain.EndpointHistoryFilter{},
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

	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result.Entries))
	}

	if result.Entries[0].Endpoint != "/api/v1/users" {
		t.Errorf("expected endpoint /api/v1/users, got %s", result.Entries[0].Endpoint)
	}

	if result.Entries[1].Method != "POST" {
		t.Errorf("expected method POST, got %s", result.Entries[1].Method)
	}
}

func TestListEndpointHistoryHandler_Empty(t *testing.T) {
	readRepo := &mockAuditReadRepository{
		endpointViews: []*domain.EndpointHistoryView{},
		endpointTotal: 0,
	}

	handler := NewListEndpointHistoryHandler(readRepo, logger.Noop())

	q := ListEndpointHistoryQuery{
		Filter: domain.EndpointHistoryFilter{},
	}

	result, err := handler.Handle(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}

	if len(result.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result.Entries))
	}
}

func TestListEndpointHistoryHandler_RepoError(t *testing.T) {
	readRepo := &mockAuditReadRepository{
		endpointErr: errors.New("database unavailable"),
	}

	handler := NewListEndpointHistoryHandler(readRepo, logger.Noop())

	q := ListEndpointHistoryQuery{
		Filter: domain.EndpointHistoryFilter{},
	}

	result, err := handler.Handle(context.Background(), q)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestListEndpointHistoryHandler_MapsAllFields(t *testing.T) {
	userID := uuid.New()
	entryID := uuid.New()
	ip := "10.0.0.1"
	ua := "TestAgent"
	now := time.Now()

	readRepo := &mockAuditReadRepository{
		endpointViews: []*domain.EndpointHistoryView{
			{
				ID:         entryID,
				UserID:     &userID,
				Endpoint:   "/api/v1/health",
				Method:     "GET",
				StatusCode: 200,
				Latency:    15,
				IPAddress:  &ip,
				UserAgent:  &ua,
				CreatedAt:  now,
			},
		},
		endpointTotal: 1,
	}

	handler := NewListEndpointHistoryHandler(readRepo, logger.Noop())

	result, err := handler.Handle(context.Background(), ListEndpointHistoryQuery{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	v := result.Entries[0]

	if v.ID != entryID {
		t.Error("ID mismatch")
	}
	if *v.UserID != userID {
		t.Error("userID mismatch")
	}
	if v.Endpoint != "/api/v1/health" {
		t.Errorf("expected endpoint /api/v1/health, got %s", v.Endpoint)
	}
	if v.Method != "GET" {
		t.Errorf("expected method GET, got %s", v.Method)
	}
	if v.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", v.StatusCode)
	}
	if v.Latency != 15 {
		t.Errorf("expected latency 15, got %d", v.Latency)
	}
	if *v.IPAddress != "10.0.0.1" {
		t.Error("ipAddress mismatch")
	}
	if *v.UserAgent != "TestAgent" {
		t.Error("userAgent mismatch")
	}
	if v.CreatedAt.IsZero() {
		t.Error("createdAt should not be zero")
	}
}
