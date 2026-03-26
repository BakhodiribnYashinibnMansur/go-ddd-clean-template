package command

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestDeletePolicyHandler_Success(t *testing.T) {
	repo := &mockPolicyRepository{}
	log := &mockLogger{}

	handler := NewDeletePolicyHandler(repo, log)

	cmd := DeletePolicyCommand{ID: uuid.New()}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}
