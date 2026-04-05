package command

import (
	"context"
	"testing"
	"time"

	"gct/internal/context/admin/supporting/errorcode/domain"

	"github.com/stretchr/testify/require"
)

func TestDeleteErrorCodeHandler_Handle(t *testing.T) {
	t.Parallel()

	id := domain.NewErrorCodeID()
	ec := domain.ReconstructErrorCode(id.UUID(), time.Now(), time.Now(), "ERR_TEST", "test", "", "", 500, "SYSTEM", "LOW", false, 0, "")
	repo := &mockErrorCodeRepo{
		findFn: func(_ context.Context, gotID domain.ErrorCodeID) (*domain.ErrorCode, error) {
			if gotID != id {
				t.Errorf("FindByID called with wrong ID: got %s, want %s", gotID, id)
			}
			return ec, nil
		},
	}
	eb := &mockEventBus{}
	handler := NewDeleteErrorCodeHandler(repo, eb, &mockLogger{})

	err := handler.Handle(context.Background(), DeleteErrorCodeCommand{ID: domain.ErrorCodeID(id)})
	require.NoError(t, err)
	if repo.deleted != id {
		t.Errorf("expected deleted ID %s, got %s", id, repo.deleted)
	}
	if len(eb.published) == 0 {
		t.Fatal("expected ErrorCodeDeleted event to be published")
	}
	if eb.published[0].EventName() != "errorcode.deleted" {
		t.Errorf("expected errorcode.deleted event, got %s", eb.published[0].EventName())
	}
}
