package command

import (
	"context"
	"errors"
	"testing"

	announceentity "gct/internal/context/content/supporting/announcement/domain/entity"
	announcerepo "gct/internal/context/content/supporting/announcement/domain/repository"
	shared "gct/internal/kernel/domain"
)

// --- Error mocks ---

var errRepoSave = errors.New("repo save failed")
var errRepoUpdate = errors.New("repo update failed")
var errRepoDelete = errors.New("repo delete failed")

type errorAnnouncementRepo struct {
	saveErr   error
	updateErr error
	deleteErr error
	findFn    func(ctx context.Context, id announceentity.AnnouncementID) (*announceentity.Announcement, error)
}

func (m *errorAnnouncementRepo) Save(_ context.Context, _ *announceentity.Announcement) error {
	return m.saveErr
}

func (m *errorAnnouncementRepo) FindByID(ctx context.Context, id announceentity.AnnouncementID) (*announceentity.Announcement, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, announceentity.ErrAnnouncementNotFound
}

func (m *errorAnnouncementRepo) Update(_ context.Context, _ *announceentity.Announcement) error {
	return m.updateErr
}

func (m *errorAnnouncementRepo) Delete(_ context.Context, _ announceentity.AnnouncementID) error {
	return m.deleteErr
}

func (m *errorAnnouncementRepo) List(_ context.Context, _ announcerepo.AnnouncementFilter) ([]*announceentity.Announcement, int64, error) {
	return nil, 0, nil
}

// --- Tests ---

func TestCreateAnnouncementHandler_RepoError(t *testing.T) {
	t.Parallel()

	repo := &errorAnnouncementRepo{saveErr: errRepoSave}
	handler := NewCreateAnnouncementHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), CreateAnnouncementCommand{
		Title:   shared.Lang{Uz: "t", Ru: "t", En: "t"},
		Content: shared.Lang{Uz: "c", Ru: "c", En: "c"},
	})
	if !errors.Is(err, errRepoSave) {
		t.Fatalf("expected errRepoSave, got: %v", err)
	}
}

func TestUpdateAnnouncementHandler_RepoUpdateError(t *testing.T) {
	t.Parallel()

	a, _ := announceentity.NewAnnouncement(
		shared.Lang{Uz: "t", Ru: "t", En: "t"},
		shared.Lang{Uz: "c", Ru: "c", En: "c"},
		1, nil, nil,
	)

	repo := &errorAnnouncementRepo{
		findFn:    func(_ context.Context, _ announceentity.AnnouncementID) (*announceentity.Announcement, error) { return a, nil },
		updateErr: errRepoUpdate,
	}
	handler := NewUpdateAnnouncementHandler(repo, &mockEventBus{}, &mockLogger{})

	err := handler.Handle(context.Background(), UpdateAnnouncementCommand{ID: announceentity.AnnouncementID(a.ID())})
	if !errors.Is(err, errRepoUpdate) {
		t.Fatalf("expected errRepoUpdate, got: %v", err)
	}
}

func TestDeleteAnnouncementHandler_RepoError(t *testing.T) {
	t.Parallel()

	repo := &errorAnnouncementRepo{deleteErr: errRepoDelete}
	handler := NewDeleteAnnouncementHandler(repo, &mockLogger{})

	err := handler.Handle(context.Background(), DeleteAnnouncementCommand{ID: announceentity.NewAnnouncementID()})
	if !errors.Is(err, errRepoDelete) {
		t.Fatalf("expected errRepoDelete, got: %v", err)
	}
}
