package command

import (
	"context"
	"errors"
	"testing"

	"gct/internal/context/admin/supporting/sitesetting/domain"

	"github.com/google/uuid"
)

// errorRepo is a mock that returns configurable errors for each method.
type errorRepo struct {
	saveErr   error
	updateErr error
	deleteErr error
	findFn    func(ctx context.Context, id domain.SiteSettingID) (*domain.SiteSetting, error)
}

func (m *errorRepo) Save(_ context.Context, _ *domain.SiteSetting) error {
	return m.saveErr
}

func (m *errorRepo) FindByID(ctx context.Context, id domain.SiteSettingID) (*domain.SiteSetting, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrSiteSettingNotFound
}

func (m *errorRepo) Update(_ context.Context, _ *domain.SiteSetting) error {
	return m.updateErr
}

func (m *errorRepo) Delete(_ context.Context, _ domain.SiteSettingID) error {
	return m.deleteErr
}

func (m *errorRepo) List(_ context.Context, _ domain.SiteSettingFilter) ([]*domain.SiteSetting, int64, error) {
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

	handler := NewCreateSiteSettingHandler(repo, eb, log)
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

	handler := NewUpdateSiteSettingHandler(repo, eb, log)
	newVal := "new"
	err := handler.Handle(context.Background(), UpdateSiteSettingCommand{
		ID:    domain.SiteSettingID(uuid.New()),
		Value: &newVal,
	})
	if err == nil {
		t.Fatal("expected error for not found")
	}
}

func TestUpdateSiteSettingHandler_UpdateError(t *testing.T) {
	t.Parallel()

	ss := domain.NewSiteSetting("k", "v", "t", "d")

	repo := &errorRepo{
		findFn:    func(_ context.Context, _ domain.SiteSettingID) (*domain.SiteSetting, error) { return ss, nil },
		updateErr: errUpdate,
	}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewUpdateSiteSettingHandler(repo, eb, log)
	newVal := "updated"
	err := handler.Handle(context.Background(), UpdateSiteSettingCommand{
		ID:    domain.SiteSettingID(ss.TypedID()),
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

	handler := NewDeleteSiteSettingHandler(repo, log)
	err := handler.Handle(context.Background(), DeleteSiteSettingCommand{ID: domain.SiteSettingID(uuid.New())})
	if !errors.Is(err, errDelete) {
		t.Fatalf("expected errDelete, got: %v", err)
	}
}
