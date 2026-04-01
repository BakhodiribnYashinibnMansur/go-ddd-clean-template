package query

import (
	"context"
	"errors"
	"testing"

	"gct/internal/integration/domain"

	"github.com/google/uuid"
)

func TestValidateAPIKeyHandler_Success(t *testing.T) {
	keyID := uuid.New()
	integrationID := uuid.New()

	readRepo := &mockReadRepo{
		apiKeyView: &domain.IntegrationAPIKeyView{
			ID:            keyID,
			IntegrationID: integrationID,
			Key:           "sk-test-key-123",
			Active:        true,
		},
	}
	l := &mockLogger{}

	handler := NewValidateAPIKeyHandler(readRepo, l)

	result, err := handler.Handle(context.Background(), ValidateAPIKeyQuery{
		APIKey: "sk-test-key-123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.ID != keyID {
		t.Errorf("expected ID %s, got %s", keyID, result.ID)
	}
	if result.IntegrationID != integrationID {
		t.Errorf("expected integration ID %s, got %s", integrationID, result.IntegrationID)
	}
	if result.Key != "sk-test-key-123" {
		t.Errorf("expected key sk-test-key-123, got %s", result.Key)
	}
	if !result.Active {
		t.Error("expected active true")
	}
}

func TestValidateAPIKeyHandler_KeyNotFound(t *testing.T) {
	readRepo := &mockReadRepo{} // FindByAPIKey returns ErrIntegrationNotFound by default
	l := &mockLogger{}

	handler := NewValidateAPIKeyHandler(readRepo, l)

	result, err := handler.Handle(context.Background(), ValidateAPIKeyQuery{
		APIKey: "nonexistent-key",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Fatalf("expected nil result on error, got %v", result)
	}
}

func TestValidateAPIKeyHandler_Inactive(t *testing.T) {
	readRepo := &mockReadRepo{
		apiKeyView: &domain.IntegrationAPIKeyView{
			ID:            uuid.New(),
			IntegrationID: uuid.New(),
			Key:           "sk-inactive-key",
			Active:        false,
		},
	}
	l := &mockLogger{}

	handler := NewValidateAPIKeyHandler(readRepo, l)

	result, err := handler.Handle(context.Background(), ValidateAPIKeyQuery{
		APIKey: "sk-inactive-key",
	})
	if err == nil {
		t.Fatal("expected error for inactive key, got nil")
	}
	if !errors.Is(err, domain.ErrAPIKeyInactive) {
		t.Fatalf("expected ErrAPIKeyInactive, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result for inactive key, got %v", result)
	}
}

func TestValidateAPIKeyHandler_RepoError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	readRepo := &errorReadRepo{err: repoErr}
	l := &mockLogger{}

	handler := NewValidateAPIKeyHandler(readRepo, l)

	result, err := handler.Handle(context.Background(), ValidateAPIKeyQuery{
		APIKey: "any-key",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result on repo error, got %v", result)
	}
}
