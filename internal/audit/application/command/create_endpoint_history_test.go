package command

import (
	"context"
	"errors"
	"testing"

	"gct/internal/audit/domain"

	"github.com/google/uuid"
)

// --- Mock EndpointHistory Repository ---

type mockEndpointHistoryRepository struct {
	savedEntry *domain.EndpointHistory
	saveErr    error
}

func (m *mockEndpointHistoryRepository) Save(_ context.Context, entry *domain.EndpointHistory) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedEntry = entry
	return nil
}

// --- Tests ---

func TestCreateEndpointHistoryHandler_Handle(t *testing.T) {
	repo := &mockEndpointHistoryRepository{}
	log := &mockLogger{}

	handler := NewCreateEndpointHistoryHandler(repo, log)

	userID := uuid.New()
	ip := "192.168.1.100"
	ua := "Mozilla/5.0"

	cmd := CreateEndpointHistoryCommand{
		UserID:     &userID,
		Endpoint:   "/api/v1/users",
		Method:     "GET",
		StatusCode: 200,
		Latency:    42,
		IPAddress:  &ip,
		UserAgent:  &ua,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.savedEntry == nil {
		t.Fatal("expected endpoint history to be saved, but it was nil")
	}

	if repo.savedEntry.Endpoint() != "/api/v1/users" {
		t.Errorf("expected endpoint /api/v1/users, got %s", repo.savedEntry.Endpoint())
	}

	if repo.savedEntry.Method() != "GET" {
		t.Errorf("expected method GET, got %s", repo.savedEntry.Method())
	}

	if repo.savedEntry.StatusCode() != 200 {
		t.Errorf("expected status code 200, got %d", repo.savedEntry.StatusCode())
	}

	if repo.savedEntry.Latency() != 42 {
		t.Errorf("expected latency 42, got %d", repo.savedEntry.Latency())
	}

	if repo.savedEntry.UserID() == nil || *repo.savedEntry.UserID() != userID {
		t.Error("expected userID to match")
	}

	if repo.savedEntry.IPAddress() == nil || *repo.savedEntry.IPAddress() != "192.168.1.100" {
		t.Error("expected ipAddress to be 192.168.1.100")
	}

	if repo.savedEntry.UserAgent() == nil || *repo.savedEntry.UserAgent() != "Mozilla/5.0" {
		t.Error("expected userAgent to be Mozilla/5.0")
	}
}

func TestCreateEndpointHistoryHandler_RepoError(t *testing.T) {
	repoErr := errors.New("db connection failed")
	repo := &mockEndpointHistoryRepository{saveErr: repoErr}
	log := &mockLogger{}

	handler := NewCreateEndpointHistoryHandler(repo, log)

	cmd := CreateEndpointHistoryCommand{
		Endpoint:   "/api/v1/health",
		Method:     "GET",
		StatusCode: 500,
		Latency:    100,
	}

	err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}

	if repo.savedEntry != nil {
		t.Error("expected no entry to be saved on error")
	}
}

func TestCreateEndpointHistoryHandler_NilOptionalFields(t *testing.T) {
	repo := &mockEndpointHistoryRepository{}
	log := &mockLogger{}

	handler := NewCreateEndpointHistoryHandler(repo, log)

	cmd := CreateEndpointHistoryCommand{
		UserID:     nil,
		Endpoint:   "/api/v1/audit-logs",
		Method:     "POST",
		StatusCode: 201,
		Latency:    15,
		IPAddress:  nil,
		UserAgent:  nil,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.savedEntry == nil {
		t.Fatal("expected entry to be saved")
	}

	if repo.savedEntry.UserID() != nil {
		t.Error("expected userID to be nil")
	}

	if repo.savedEntry.IPAddress() != nil {
		t.Error("expected ipAddress to be nil")
	}

	if repo.savedEntry.UserAgent() != nil {
		t.Error("expected userAgent to be nil")
	}
}

func TestCreateEndpointHistoryHandler_VariousStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		method     string
	}{
		{"404 not found", 404, "GET"},
		{"401 unauthorized", 401, "POST"},
		{"403 forbidden", 403, "DELETE"},
		{"500 server error", 500, "PATCH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockEndpointHistoryRepository{}
			log := &mockLogger{}
			handler := NewCreateEndpointHistoryHandler(repo, log)

			cmd := CreateEndpointHistoryCommand{
				Endpoint:   "/api/v1/test",
				Method:     tt.method,
				StatusCode: tt.statusCode,
				Latency:    50,
			}

			err := handler.Handle(context.Background(), cmd)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if repo.savedEntry.StatusCode() != tt.statusCode {
				t.Errorf("expected status code %d, got %d", tt.statusCode, repo.savedEntry.StatusCode())
			}

			if repo.savedEntry.Method() != tt.method {
				t.Errorf("expected method %s, got %s", tt.method, repo.savedEntry.Method())
			}
		})
	}
}
