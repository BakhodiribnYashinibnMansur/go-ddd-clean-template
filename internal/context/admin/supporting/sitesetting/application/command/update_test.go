package command

import (
	"context"
	"errors"
	"testing"

	siteentity "gct/internal/context/admin/supporting/sitesetting/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUpdateSiteSettingHandler_Handle(t *testing.T) {
	t.Parallel()

	ss := siteentity.NewSiteSetting("old_key", "old_value", "general", "old desc")

	repo := &mockRepo{
		findFn: func(_ context.Context, id siteentity.SiteSettingID) (*siteentity.SiteSetting, error) {
			if id == ss.TypedID() {
				return ss, nil
			}
			return nil, siteentity.ErrSiteSettingNotFound
		},
	}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateSiteSettingHandler(repo, eb, log)

	newValue := "new_value"
	newDesc := "new desc"
	cmd := UpdateSiteSettingCommand{
		ID:          siteentity.SiteSettingID(ss.TypedID()),
		Value:       &newValue,
		Description: &newDesc,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.updated == nil {
		t.Fatal("expected site setting to be updated")
	}
	if repo.updated.Value() != "new_value" {
		t.Errorf("expected value new_value, got %s", repo.updated.Value())
	}
	if repo.updated.Description() != "new desc" {
		t.Errorf("expected description new desc, got %s", repo.updated.Description())
	}
	// Unchanged fields should be preserved
	if repo.updated.Key() != "old_key" {
		t.Errorf("expected key old_key (unchanged), got %s", repo.updated.Key())
	}
	if repo.updated.Type() != "general" {
		t.Errorf("expected type general (unchanged), got %s", repo.updated.Type())
	}

	if len(eb.published) == 0 {
		t.Fatal("expected events to be published")
	}
	if eb.published[0].EventName() != "sitesetting.updated" {
		t.Errorf("expected sitesetting.updated event, got %s", eb.published[0].EventName())
	}
}

func TestUpdateSiteSettingHandler_NotFound(t *testing.T) {
	t.Parallel()

	repo := &mockRepo{}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateSiteSettingHandler(repo, eb, log)

	newVal := "v"
	err := handler.Handle(context.Background(), UpdateSiteSettingCommand{
		ID:    siteentity.SiteSettingID(uuid.New()),
		Value: &newVal,
	})
	if err == nil {
		t.Fatal("expected error for non-existent site setting")
	}
}

func TestUpdateSiteSettingHandler_RepoUpdateError(t *testing.T) {
	t.Parallel()

	ss := siteentity.NewSiteSetting("k", "v", "t", "d")
	repoErr := errors.New("repo update failed")

	errR := &errorRepo{
		findFn:    func(_ context.Context, _ siteentity.SiteSettingID) (*siteentity.SiteSetting, error) { return ss, nil },
		updateErr: repoErr,
	}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateSiteSettingHandler(errR, eb, log)

	newVal := "new"
	err := handler.Handle(context.Background(), UpdateSiteSettingCommand{
		ID:    siteentity.SiteSettingID(ss.TypedID()),
		Value: &newVal,
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo update error, got: %v", err)
	}
}
