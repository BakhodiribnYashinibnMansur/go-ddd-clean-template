package query

import (
	"context"
	"errors"
	"testing"

	siteentity "gct/internal/context/admin/supporting/sitesetting/domain/entity"
	siterepo "gct/internal/context/admin/supporting/sitesetting/domain/repository"
	"gct/internal/kernel/infrastructure/logger"
)

// mockMaxSessionsRepo is a minimal read-repo double that just echoes a
// preset List result. Keeping it local avoids leaking the mock surface
// used by the other query tests.
type mockMaxSessionsRepo struct {
	views []*siterepo.SiteSettingView
	err   error
}

func (m *mockMaxSessionsRepo) FindByID(_ context.Context, _ siteentity.SiteSettingID) (*siterepo.SiteSettingView, error) {
	return nil, siteentity.ErrSiteSettingNotFound
}

func (m *mockMaxSessionsRepo) List(_ context.Context, _ siterepo.SiteSettingFilter) ([]*siterepo.SiteSettingView, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	return m.views, int64(len(m.views)), nil
}

func TestGetUserMaxSessions_DefaultWhenMissing(t *testing.T) {
	t.Parallel()
	h := NewGetUserMaxSessionsHandler(&mockMaxSessionsRepo{}, logger.Noop())
	got, err := h.Handle(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != DefaultUserMaxSessions {
		t.Errorf("expected %d, got %d", DefaultUserMaxSessions, got)
	}
}

func TestGetUserMaxSessions_ParsesValue(t *testing.T) {
	t.Parallel()
	repo := &mockMaxSessionsRepo{views: []*siterepo.SiteSettingView{{Key: UserMaxSessionsKey, Value: "7"}}}
	h := NewGetUserMaxSessionsHandler(repo, logger.Noop())
	got, err := h.Handle(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != 7 {
		t.Errorf("expected 7, got %d", got)
	}
}

func TestGetUserMaxSessions_MalformedFallsBack(t *testing.T) {
	t.Parallel()
	repo := &mockMaxSessionsRepo{views: []*siterepo.SiteSettingView{{Key: UserMaxSessionsKey, Value: "not-an-int"}}}
	h := NewGetUserMaxSessionsHandler(repo, logger.Noop())
	got, err := h.Handle(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != DefaultUserMaxSessions {
		t.Errorf("expected default %d, got %d", DefaultUserMaxSessions, got)
	}
}

func TestGetUserMaxSessions_NonPositiveFallsBack(t *testing.T) {
	t.Parallel()
	repo := &mockMaxSessionsRepo{views: []*siterepo.SiteSettingView{{Key: UserMaxSessionsKey, Value: "0"}}}
	h := NewGetUserMaxSessionsHandler(repo, logger.Noop())
	got, err := h.Handle(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != DefaultUserMaxSessions {
		t.Errorf("expected default %d, got %d", DefaultUserMaxSessions, got)
	}
}

func TestGetUserMaxSessions_RepoErrorDegrades(t *testing.T) {
	t.Parallel()
	repo := &mockMaxSessionsRepo{err: errors.New("boom")}
	h := NewGetUserMaxSessionsHandler(repo, logger.Noop())
	got, err := h.Handle(context.Background())
	if err != nil {
		t.Fatalf("handler must never surface repo errors; got %v", err)
	}
	if got != DefaultUserMaxSessions {
		t.Errorf("expected default %d, got %d", DefaultUserMaxSessions, got)
	}
}
