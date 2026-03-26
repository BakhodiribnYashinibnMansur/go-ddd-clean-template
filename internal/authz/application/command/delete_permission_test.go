package command

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestDeletePermissionHandler_Success(t *testing.T) {
	repo := &mockPermissionRepository{}
	log := &mockLogger{}

	handler := NewDeletePermissionHandler(repo, log)

	cmd := DeletePermissionCommand{ID: uuid.New()}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}
