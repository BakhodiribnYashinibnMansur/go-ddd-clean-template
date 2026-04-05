package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/context/content/announcement"
	"gct/internal/context/content/announcement/application/command"
	"gct/internal/context/content/announcement/application/query"
	"gct/internal/context/content/announcement/domain"
	"gct/internal/platform/application"
	shared "gct/internal/platform/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- Mocks ---

type mockRepo struct {
	saved   *domain.Announcement
	updated *domain.Announcement
	deleted uuid.UUID
	findFn  func(ctx context.Context, id uuid.UUID) (*domain.Announcement, error)
}

func (m *mockRepo) Save(_ context.Context, e *domain.Announcement) error {
	m.saved = e
	return nil
}
func (m *mockRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Announcement, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrAnnouncementNotFound
}
func (m *mockRepo) Update(_ context.Context, e *domain.Announcement) error {
	m.updated = e
	return nil
}
func (m *mockRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.deleted = id
	return nil
}
func (m *mockRepo) List(_ context.Context, _ domain.AnnouncementFilter) ([]*domain.Announcement, int64, error) {
	return nil, 0, nil
}

type mockReadRepo struct {
	view  *domain.AnnouncementView
	views []*domain.AnnouncementView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.AnnouncementView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrAnnouncementNotFound
}
func (m *mockReadRepo) List(_ context.Context, _ domain.AnnouncementFilter) ([]*domain.AnnouncementView, int64, error) {
	return m.views, m.total, nil
}

type mockEventBus struct{ published []shared.DomainEvent }

func (m *mockEventBus) Publish(_ context.Context, events ...shared.DomainEvent) error {
	m.published = append(m.published, events...)
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

// --- Helpers ---

func setupRouter(repo *mockRepo, readRepo *mockReadRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)

	eb := &mockEventBus{}
	l := &mockLogger{}

	bc := &announcement.BoundedContext{
		CreateAnnouncement: command.NewCreateAnnouncementHandler(repo, eb, l),
		UpdateAnnouncement: command.NewUpdateAnnouncementHandler(repo, eb, l),
		DeleteAnnouncement: command.NewDeleteAnnouncementHandler(repo, l),
		GetAnnouncement:    query.NewGetAnnouncementHandler(readRepo, l),
		ListAnnouncements:  query.NewListAnnouncementsHandler(readRepo, l),
	}

	r := gin.New()
	h := NewHandler(bc, l)
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	return r
}

// --- Tests ---

func TestHandler_Create_Success(t *testing.T) {
	repo := &mockRepo{}
	readRepo := &mockReadRepo{}
	router := setupRouter(repo, readRepo)

	body := CreateRequest{
		Title:   shared.Lang{Uz: "t_uz", Ru: "t_ru", En: "t_en"},
		Content: shared.Lang{Uz: "c_uz", Ru: "c_ru", En: "c_en"},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/announcements", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Create_BadRequest(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/announcements", bytes.NewBufferString(`not json`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_List_Success(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*domain.AnnouncementView{
			{ID: uuid.New(), TitleEn: "A1", Priority: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		total: 1,
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/announcements?limit=10&offset=0", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_Success(t *testing.T) {
	id := uuid.New()
	readRepo := &mockReadRepo{
		view: &domain.AnnouncementView{
			ID: id, TitleEn: "A", Priority: 1, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/announcements/"+id.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_InvalidID(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/announcements/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Update_Success(t *testing.T) {
	a := domain.NewAnnouncement(
		shared.Lang{Uz: "t", Ru: "t", En: "t"},
		shared.Lang{Uz: "c", Ru: "c", En: "c"},
		1, nil, nil,
	)
	repo := &mockRepo{
		findFn: func(_ context.Context, id uuid.UUID) (*domain.Announcement, error) {
			if id == a.ID() {
				return a, nil
			}
			return nil, domain.ErrAnnouncementNotFound
		},
	}
	router := setupRouter(repo, &mockReadRepo{})

	body := UpdateRequest{Publish: true}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/announcements/"+a.ID().String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Update_InvalidID(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/announcements/bad-id", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Delete_Success(t *testing.T) {
	repo := &mockRepo{}
	router := setupRouter(repo, &mockReadRepo{})

	id := uuid.New()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/announcements/"+id.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/announcements/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
