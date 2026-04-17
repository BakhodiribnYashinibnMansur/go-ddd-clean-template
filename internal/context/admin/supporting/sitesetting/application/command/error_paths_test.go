package command

import (
	"context"
	"errors"
	"testing"

	siteentity "gct/internal/context/admin/supporting/sitesetting/domain/entity"
	siterepo "gct/internal/context/admin/supporting/sitesetting/domain/repository"
	shareddomain "gct/internal/kernel/domain"

	"gct/internal/kernel/outbox"
	"github.com/google/uuid"
)

// errorRepo is a mock that returns configurable errors for each method.
type errorRepo struct {
	saveErr   error
	updateErr error
	deleteErr error
	findFn    func(ctx context.Context, id siteentity.SiteSettingID) (*siteentity.SiteSetting, error)
}

func (m *errorRepo) Save(_ context.Context, _ shareddomain.Querier, _ *siteentity.SiteSetting) error {
	return m.saveErr
}

func (m *errorRepo) FindByID(ctx context.Context, id siteentity.SiteSettingID) (*siteentity.SiteSetting, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, siteentity.ErrSiteSettingNotFound
}

func (m *errorRepo) Update(_ context.Context, _ shareddomain.Querier, _ *siteentity.SiteSetting) error {
	return m.updateErr
}

func (m *errorRepo) Delete(_ context.Context, _ shareddomain.Querier, _ siteentity.SiteSettingID) error {
	return m.deleteErr
}

func (m *errorRepo) List(_ context.Context, _ siterepo.SiteSettingFilter) ([]*siteentity.SiteSetting, int64, error) {
	return nil, 0, nil
}

// --- Error path tests ---

var errSave = errors.New("save failed")
var errUpdate = errors.New("update failed")
var errDelete = errors.New("delete failed")

func TestCreateSiteSettingHandler_SaveError(t *testing.T) {
	t.Parallel()

	repo := &errorRepo{saveErr: errSave}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateSiteSettingHandler(repo, outbox.NewEventCommitter(nil, nil, eb, log), log)
	err := handler.Handle(context.Background(), CreateSiteSettingCommand{
		Key: "k", Value: "v", Type: "t",
	})
	if !errors.Is(err, errSave) {
		t.Fatalf("expected errSave, got: %v", err)
	}
}

func TestUpdateSiteSettingHandler_FindError(t *testing.T) {
	t.Parallel()

	repo := &errorRepo{}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateSiteSettingHandler(repo, outbox.NewEventCommitter(nil, nil, eb, log), log)
	newVal := "new"
	err := handler.Handle(context.Background(), UpdateSiteSettingCommand{
		ID:    siteentity.SiteSettingID(uuid.New()),
		Value: &newVal,
	})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestUpdateSiteSettingHandler_UpdateError(t *testing.T) {
	t.Parallel()

	ss := siteentity.NewSiteSetting("k", "v", "t", "d")

	repo := &errorRepo{
		findFn:    func(_ context.Context, _ siteentity.SiteSettingID) (*siteentity.SiteSetting, error) { return ss, nil },
		updateErr: errUpdate,
	}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateSiteSettingHandler(repo, outbox.NewEventCommitter(nil, nil, eb, log), log)
	newVal := "updated"
	err := handler.Handle(context.Background(), UpdateSiteSettingCommand{
		ID:    siteentity.SiteSettingID(ss.TypedID()),
		Value: &newVal,
	})
	if !errors.Is(err, errUpdate) {
		t.Fatalf("expected errUpdate, got: %v", err)
	}
}

func TestDeleteSiteSettingHandler_DeleteError(t *testing.T) {
	t.Parallel()

	repo := &errorRepo{deleteErr: errDelete}
	log := &mockLogger{}

	handler := NewDeleteSiteSettingHandler(repo, nil, log)
	err := handler.Handle(context.Background(), DeleteSiteSettingCommand{ID: siteentity.SiteSettingID(uuid.New())})
	if !errors.Is(err, errDelete) {
		t.Fatalf("expected errDelete, got: %v", err)
	}
}
