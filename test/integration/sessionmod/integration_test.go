package sessionmod

import (
	"context"
	"testing"

	"gct/internal/context/iam/session"
	appdto "gct/internal/context/iam/session/application"
	"gct/internal/context/iam/session/application/query"
	"gct/internal/platform/infrastructure/logger"
	"gct/test/integration/common/setup"
)

func newTestBC(t *testing.T) *session.BoundedContext {
	t.Helper()
	l := logger.New("error")
	return session.NewBoundedContext(setup.TestPG.Pool, l)
}

func TestIntegration_ListSessions_Empty(t *testing.T) {
	cleanDB(t)
	bc := newTestBC(t)
	ctx := context.Background()

	result, err := bc.ListSessions.Handle(ctx, query.ListSessionsQuery{
		Filter: appdto.SessionsFilter{Limit: 10},
	})
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected 0 sessions, got %d", result.Total)
	}
	if len(result.Sessions) != 0 {
		t.Errorf("expected empty sessions slice, got %d items", len(result.Sessions))
	}
}
