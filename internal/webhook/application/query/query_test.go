package query

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/webhook/domain"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Mock infrastructure
// ---------------------------------------------------------------------------

type mockReadRepo struct {
	view  *domain.WebhookView
	views []*domain.WebhookView
	total int64

	findByIDFn func(ctx context.Context, id uuid.UUID) (*domain.WebhookView, error)
	listFn     func(ctx context.Context, filter domain.WebhookFilter) ([]*domain.WebhookView, int64, error)
}

func (m *mockReadRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.WebhookView, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrWebhookNotFound
}

func (m *mockReadRepo) List(ctx context.Context, filter domain.WebhookFilter) ([]*domain.WebhookView, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return m.views, m.total, nil
}

// ---------------------------------------------------------------------------
// Tests: GetHandler
// ---------------------------------------------------------------------------

func TestGetHandler_Success(t *testing.T) {
	webhookID := uuid.New()
	now := time.Now()

	readRepo := &mockReadRepo{
		view: &domain.WebhookView{
			ID:        webhookID,
			Name:      "my-hook",
			URL:       "https://example.com/hook",
			Secret:    "s3cret",
			Events:    []string{"user.created"},
			Enabled:   true,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	h := NewGetHandler(readRepo)

	result, err := h.Handle(context.Background(), GetQuery{ID: webhookID})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.ID != webhookID {
		t.Fatalf("expected ID %s, got %s", webhookID, result.ID)
	}
	if result.Name != "my-hook" {
		t.Fatalf("expected name my-hook, got %s", result.Name)
	}
	if result.URL != "https://example.com/hook" {
		t.Fatalf("expected url https://example.com/hook, got %s", result.URL)
	}
	if result.Secret != "s3cret" {
		t.Fatalf("expected secret s3cret, got %s", result.Secret)
	}
	if len(result.Events) != 1 || result.Events[0] != "user.created" {
		t.Fatalf("expected events [user.created], got %v", result.Events)
	}
	if !result.Enabled {
		t.Fatal("expected enabled true")
	}
}

func TestGetHandler_NotFound(t *testing.T) {
	readRepo := &mockReadRepo{} // no view set
	h := NewGetHandler(readRepo)

	_, err := h.Handle(context.Background(), GetQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrWebhookNotFound) {
		t.Fatalf("expected ErrWebhookNotFound, got %v", err)
	}
}

func TestGetHandler_RepoError(t *testing.T) {
	repoErr := errors.New("db timeout")
	readRepo := &mockReadRepo{
		findByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.WebhookView, error) {
			return nil, repoErr
		},
	}
	h := NewGetHandler(readRepo)

	_, err := h.Handle(context.Background(), GetQuery{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != repoErr {
		t.Fatalf("expected repo error, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Tests: ListHandler
// ---------------------------------------------------------------------------

func TestListHandler_Success(t *testing.T) {
	now := time.Now()
	id1 := uuid.New()
	id2 := uuid.New()

	readRepo := &mockReadRepo{
		views: []*domain.WebhookView{
			{
				ID:        id1,
				Name:      "hook-1",
				URL:       "https://example.com/1",
				Secret:    "sec1",
				Events:    []string{"user.created"},
				Enabled:   true,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				ID:        id2,
				Name:      "hook-2",
				URL:       "https://example.com/2",
				Secret:    "sec2",
				Events:    []string{"user.deleted"},
				Enabled:   false,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		total: 2,
	}
	h := NewListHandler(readRepo)

	q := ListQuery{
		Filter: domain.WebhookFilter{Limit: 10, Offset: 0},
	}
	result, err := h.Handle(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Total != 2 {
		t.Fatalf("expected total 2, got %d", result.Total)
	}
	if len(result.Webhooks) != 2 {
		t.Fatalf("expected 2 webhooks, got %d", len(result.Webhooks))
	}
	if result.Webhooks[0].Name != "hook-1" {
		t.Fatalf("expected first webhook name hook-1, got %s", result.Webhooks[0].Name)
	}
	if result.Webhooks[1].Name != "hook-2" {
		t.Fatalf("expected second webhook name hook-2, got %s", result.Webhooks[1].Name)
	}
}

func TestListHandler_Empty(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*domain.WebhookView{},
		total: 0,
	}
	h := NewListHandler(readRepo)

	q := ListQuery{
		Filter: domain.WebhookFilter{Limit: 10, Offset: 0},
	}
	result, err := h.Handle(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Total != 0 {
		t.Fatalf("expected total 0, got %d", result.Total)
	}
	if len(result.Webhooks) != 0 {
		t.Fatalf("expected 0 webhooks, got %d", len(result.Webhooks))
	}
}

func TestListHandler_RepoError(t *testing.T) {
	repoErr := errors.New("db connection lost")
	readRepo := &mockReadRepo{
		listFn: func(_ context.Context, _ domain.WebhookFilter) ([]*domain.WebhookView, int64, error) {
			return nil, 0, repoErr
		},
	}
	h := NewListHandler(readRepo)

	q := ListQuery{
		Filter: domain.WebhookFilter{Limit: 10, Offset: 0},
	}
	_, err := h.Handle(context.Background(), q)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != repoErr {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestListHandler_WithFilter(t *testing.T) {
	now := time.Now()
	readRepo := &mockReadRepo{
		listFn: func(_ context.Context, filter domain.WebhookFilter) ([]*domain.WebhookView, int64, error) {
			// Verify filter is passed through
			if filter.Limit != 5 {
				return nil, 0, errors.New("expected limit 5")
			}
			if filter.Offset != 10 {
				return nil, 0, errors.New("expected offset 10")
			}
			return []*domain.WebhookView{
				{
					ID:        uuid.New(),
					Name:      "filtered-hook",
					URL:       "https://example.com",
					Events:    []string{},
					Enabled:   true,
					CreatedAt: now,
					UpdatedAt: now,
				},
			}, 1, nil
		},
	}
	h := NewListHandler(readRepo)

	q := ListQuery{
		Filter: domain.WebhookFilter{Limit: 5, Offset: 10},
	}
	result, err := h.Handle(context.Background(), q)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected total 1, got %d", result.Total)
	}
	if result.Webhooks[0].Name != "filtered-hook" {
		t.Fatalf("expected filtered-hook, got %s", result.Webhooks[0].Name)
	}
}

func TestListHandler_MapsAllFields(t *testing.T) {
	now := time.Now()
	id := uuid.New()

	readRepo := &mockReadRepo{
		views: []*domain.WebhookView{
			{
				ID:        id,
				Name:      "full-hook",
				URL:       "https://full.example.com",
				Secret:    "topsecret",
				Events:    []string{"a.b", "c.d"},
				Enabled:   false,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		total: 1,
	}
	h := NewListHandler(readRepo)

	result, err := h.Handle(context.Background(), ListQuery{Filter: domain.WebhookFilter{Limit: 10}})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	w := result.Webhooks[0]
	if w.ID != id {
		t.Fatalf("expected ID %s, got %s", id, w.ID)
	}
	if w.Name != "full-hook" {
		t.Fatalf("expected name full-hook, got %s", w.Name)
	}
	if w.URL != "https://full.example.com" {
		t.Fatalf("expected url https://full.example.com, got %s", w.URL)
	}
	if w.Secret != "topsecret" {
		t.Fatalf("expected secret topsecret, got %s", w.Secret)
	}
	if len(w.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(w.Events))
	}
	if w.Enabled {
		t.Fatal("expected enabled false")
	}
	if !w.CreatedAt.Equal(now) {
		t.Fatalf("expected CreatedAt %v, got %v", now, w.CreatedAt)
	}
	if !w.UpdatedAt.Equal(now) {
		t.Fatalf("expected UpdatedAt %v, got %v", now, w.UpdatedAt)
	}
}
