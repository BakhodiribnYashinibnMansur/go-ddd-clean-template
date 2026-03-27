package command

import (
	"context"
	"errors"
	"testing"

	"gct/internal/shared/application"
	shared "gct/internal/shared/domain"
	"gct/internal/webhook/domain"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Mock infrastructure
// ---------------------------------------------------------------------------

type mockWebhookRepo struct {
	savedWebhook   *domain.Webhook
	updatedWebhook *domain.Webhook
	deletedID      uuid.UUID
	findByIDFn     func(ctx context.Context, id uuid.UUID) (*domain.Webhook, error)
	saveFn         func(ctx context.Context, entity *domain.Webhook) error
	updateFn       func(ctx context.Context, entity *domain.Webhook) error
	deleteFn       func(ctx context.Context, id uuid.UUID) error
}

func (m *mockWebhookRepo) Save(ctx context.Context, entity *domain.Webhook) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, entity)
	}
	m.savedWebhook = entity
	return nil
}

func (m *mockWebhookRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Webhook, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, domain.ErrWebhookNotFound
}

func (m *mockWebhookRepo) Update(ctx context.Context, entity *domain.Webhook) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, entity)
	}
	m.updatedWebhook = entity
	return nil
}

func (m *mockWebhookRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	m.deletedID = id
	return nil
}

type mockEventBus struct {
	publishedEvents []shared.DomainEvent
	publishFn       func(ctx context.Context, events ...shared.DomainEvent) error
}

func (m *mockEventBus) Publish(ctx context.Context, events ...shared.DomainEvent) error {
	if m.publishFn != nil {
		return m.publishFn(ctx, events...)
	}
	m.publishedEvents = append(m.publishedEvents, events...)
	return nil
}

func (m *mockEventBus) Subscribe(_ string, _ application.EventHandler) error { return nil }

type mockLogger struct{}

func (m *mockLogger) Debug(args ...any)                                          {}
func (m *mockLogger) Debugf(template string, args ...any)                        {}
func (m *mockLogger) Debugw(msg string, keysAndValues ...any)                    {}
func (m *mockLogger) Info(args ...any)                                           {}
func (m *mockLogger) Infof(template string, args ...any)                         {}
func (m *mockLogger) Infow(msg string, keysAndValues ...any)                     {}
func (m *mockLogger) Warn(args ...any)                                           {}
func (m *mockLogger) Warnf(template string, args ...any)                         {}
func (m *mockLogger) Warnw(msg string, keysAndValues ...any)                     {}
func (m *mockLogger) Error(args ...any)                                          {}
func (m *mockLogger) Errorf(template string, args ...any)                        {}
func (m *mockLogger) Errorw(msg string, keysAndValues ...any)                    {}
func (m *mockLogger) Fatal(args ...any)                                          {}
func (m *mockLogger) Fatalf(template string, args ...any)                        {}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any)                    {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any)               {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)                {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)                {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any)               {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any)               {}

// ---------------------------------------------------------------------------
// Tests: CreateHandler
// ---------------------------------------------------------------------------

func TestCreateHandler_Success(t *testing.T) {
	repo := &mockWebhookRepo{}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := NewCreateHandler(repo, eb, l)

	cmd := CreateCommand{
		Name:    "my-hook",
		URL:     "https://example.com/hook",
		Secret:  "s3cret",
		Events:  []string{"user.created", "user.deleted"},
		Enabled: true,
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.savedWebhook == nil {
		t.Fatal("expected webhook to be saved")
	}
	if repo.savedWebhook.Name() != "my-hook" {
		t.Fatalf("expected name my-hook, got %s", repo.savedWebhook.Name())
	}
	if repo.savedWebhook.URL() != "https://example.com/hook" {
		t.Fatalf("expected url https://example.com/hook, got %s", repo.savedWebhook.URL())
	}
	if repo.savedWebhook.Secret() != "s3cret" {
		t.Fatalf("expected secret s3cret, got %s", repo.savedWebhook.Secret())
	}
	if len(repo.savedWebhook.Events_()) != 2 {
		t.Fatalf("expected 2 events, got %d", len(repo.savedWebhook.Events_()))
	}
	if !repo.savedWebhook.Enabled() {
		t.Fatal("expected enabled true")
	}
}

func TestCreateHandler_NilEvents(t *testing.T) {
	repo := &mockWebhookRepo{}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := NewCreateHandler(repo, eb, l)

	cmd := CreateCommand{
		Name:    "hook",
		URL:     "https://example.com",
		Secret:  "secret",
		Events:  nil,
		Enabled: false,
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.savedWebhook == nil {
		t.Fatal("expected webhook to be saved")
	}
	if len(repo.savedWebhook.Events_()) != 0 {
		t.Fatalf("expected 0 events for nil input, got %d", len(repo.savedWebhook.Events_()))
	}
}

func TestCreateHandler_RepoSaveError(t *testing.T) {
	repoErr := errors.New("db connection failed")
	repo := &mockWebhookRepo{
		saveFn: func(_ context.Context, _ *domain.Webhook) error {
			return repoErr
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := NewCreateHandler(repo, eb, l)

	cmd := CreateCommand{
		Name:    "hook",
		URL:     "https://example.com",
		Secret:  "secret",
		Events:  []string{"user.created"},
		Enabled: true,
	}

	err := h.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != repoErr {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestCreateHandler_EventBusError(t *testing.T) {
	repo := &mockWebhookRepo{}
	eb := &mockEventBus{
		publishFn: func(_ context.Context, _ ...shared.DomainEvent) error {
			return errors.New("event bus down")
		},
	}
	l := &mockLogger{}
	h := NewCreateHandler(repo, eb, l)

	cmd := CreateCommand{
		Name:    "hook",
		URL:     "https://example.com",
		Secret:  "secret",
		Events:  []string{"user.created"},
		Enabled: true,
	}

	// Event bus errors are logged but do not fail the command
	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error (event bus failure is non-fatal), got %v", err)
	}
	if repo.savedWebhook == nil {
		t.Fatal("expected webhook to be saved even when event bus fails")
	}
}

// ---------------------------------------------------------------------------
// Tests: UpdateHandler
// ---------------------------------------------------------------------------

func TestUpdateHandler_Success(t *testing.T) {
	existing := domain.NewWebhook("old-hook", "https://old.com", "oldsecret", []string{"a"}, true)
	existingID := existing.ID()

	repo := &mockWebhookRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Webhook, error) {
			if id == existingID {
				return existing, nil
			}
			return nil, domain.ErrWebhookNotFound
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := NewUpdateHandler(repo, eb, l)

	newName := "new-hook"
	newURL := "https://new.com"
	cmd := UpdateCommand{
		ID:   existingID,
		Name: &newName,
		URL:  &newURL,
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedWebhook == nil {
		t.Fatal("expected webhook to be updated")
	}
	if repo.updatedWebhook.Name() != "new-hook" {
		t.Fatalf("expected name new-hook, got %s", repo.updatedWebhook.Name())
	}
	if repo.updatedWebhook.URL() != "https://new.com" {
		t.Fatalf("expected url https://new.com, got %s", repo.updatedWebhook.URL())
	}
	// Unchanged fields should be preserved
	if repo.updatedWebhook.Secret() != "oldsecret" {
		t.Fatalf("expected secret oldsecret, got %s", repo.updatedWebhook.Secret())
	}
}

func TestUpdateHandler_PartialUpdate_EnabledOnly(t *testing.T) {
	existing := domain.NewWebhook("hook", "https://example.com", "secret", []string{"a"}, true)
	existingID := existing.ID()

	repo := &mockWebhookRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Webhook, error) {
			if id == existingID {
				return existing, nil
			}
			return nil, domain.ErrWebhookNotFound
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := NewUpdateHandler(repo, eb, l)

	disabled := false
	cmd := UpdateCommand{
		ID:      existingID,
		Enabled: &disabled,
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedWebhook == nil {
		t.Fatal("expected webhook to be updated")
	}
	if repo.updatedWebhook.Enabled() {
		t.Fatal("expected enabled false after partial update")
	}
	if repo.updatedWebhook.Name() != "hook" {
		t.Fatalf("expected name unchanged, got %s", repo.updatedWebhook.Name())
	}
}

func TestUpdateHandler_NotFound(t *testing.T) {
	repo := &mockWebhookRepo{} // default returns ErrWebhookNotFound
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := NewUpdateHandler(repo, eb, l)

	cmd := UpdateCommand{ID: uuid.New()}

	err := h.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrWebhookNotFound) {
		t.Fatalf("expected ErrWebhookNotFound, got %v", err)
	}
}

func TestUpdateHandler_RepoUpdateError(t *testing.T) {
	existing := domain.NewWebhook("hook", "https://example.com", "secret", []string{"a"}, true)
	existingID := existing.ID()
	repoErr := errors.New("update failed")

	repo := &mockWebhookRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Webhook, error) {
			if id == existingID {
				return existing, nil
			}
			return nil, domain.ErrWebhookNotFound
		},
		updateFn: func(_ context.Context, _ *domain.Webhook) error {
			return repoErr
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := NewUpdateHandler(repo, eb, l)

	newName := "new-name"
	cmd := UpdateCommand{
		ID:   existingID,
		Name: &newName,
	}

	err := h.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != repoErr {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestUpdateHandler_EventBusError(t *testing.T) {
	existing := domain.NewWebhook("hook", "https://example.com", "secret", []string{"a"}, true)
	existingID := existing.ID()

	repo := &mockWebhookRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Webhook, error) {
			if id == existingID {
				return existing, nil
			}
			return nil, domain.ErrWebhookNotFound
		},
	}
	eb := &mockEventBus{
		publishFn: func(_ context.Context, _ ...shared.DomainEvent) error {
			return errors.New("event bus down")
		},
	}
	l := &mockLogger{}
	h := NewUpdateHandler(repo, eb, l)

	newName := "updated"
	cmd := UpdateCommand{
		ID:   existingID,
		Name: &newName,
	}

	// Event bus errors are logged but do not fail the command
	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error (event bus failure is non-fatal), got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Tests: DeleteHandler
// ---------------------------------------------------------------------------

func TestDeleteHandler_Success(t *testing.T) {
	targetID := uuid.New()
	repo := &mockWebhookRepo{}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := NewDeleteHandler(repo, eb, l)

	cmd := DeleteCommand{ID: targetID}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.deletedID != targetID {
		t.Fatalf("expected deleted ID %s, got %s", targetID, repo.deletedID)
	}
}

func TestDeleteHandler_RepoError(t *testing.T) {
	repoErr := errors.New("delete failed")
	repo := &mockWebhookRepo{
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return repoErr
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := NewDeleteHandler(repo, eb, l)

	cmd := DeleteCommand{ID: uuid.New()}

	err := h.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != repoErr {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestDeleteHandler_NotFoundError(t *testing.T) {
	repo := &mockWebhookRepo{
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return domain.ErrWebhookNotFound
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := NewDeleteHandler(repo, eb, l)

	cmd := DeleteCommand{ID: uuid.New()}

	err := h.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrWebhookNotFound) {
		t.Fatalf("expected ErrWebhookNotFound, got %v", err)
	}
}
