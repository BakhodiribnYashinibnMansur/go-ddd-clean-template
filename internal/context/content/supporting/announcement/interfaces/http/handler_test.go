package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/context/content/supporting/announcement"
	"gct/internal/context/content/supporting/announcement/application/command"
	"gct/internal/context/content/supporting/announcement/application/query"
	announceentity "gct/internal/context/content/supporting/announcement/domain/entity"
	announcerepo "gct/internal/context/content/supporting/announcement/domain/repository"
	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- Mocks ---

type mockRepo struct {
	saved   *announceentity.Announcement
	updated *announceentity.Announcement
	deleted announceentity.AnnouncementID
	findFn  func(ctx context.Context, id announceentity.AnnouncementID) (*announceentity.Announcement, error)
}

func (m *mockRepo) Save(_ context.Context, e *announceentity.Announcement) error {
	m.saved = e
	return nil
}
func (m *mockRepo) FindByID(ctx context.Context, id announceentity.AnnouncementID) (*announceentity.Announcement, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, announceentity.ErrAnnouncementNotFound
}
func (m *mockRepo) Update(_ context.Context, e *announceentity.Announcement) error {
	m.updated = e
	return nil
}
func (m *mockRepo) Delete(_ context.Context, id announceentity.AnnouncementID) error {
	m.deleted = id
	return nil
}
func (m *mockRepo) List(_ context.Context, _ announcerepo.AnnouncementFilter) ([]*announceentity.Announcement, int64, error) {
	return nil, 0, nil
}

type mockReadRepo struct {
	view  *announcerepo.AnnouncementView
	views []*announcerepo.AnnouncementView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id announceentity.AnnouncementID) (*announcerepo.AnnouncementView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, announceentity.ErrAnnouncementNotFound
}
func (m *mockReadRepo) List(_ context.Context, _ announcerepo.AnnouncementFilter) ([]*announcerepo.AnnouncementView, int64, error) {
	return m.views, m.total, nil
}

type mockEventBus struct{ published []shared.DomainEvent }

func (m *mockEventBus) Publish(_ context.Context, events ...shared.DomainEvent) error {
	m.published = append(m.published, events...)
	return nil
}
func (m *mockEventBus) Subscribe(_ string, _ application.EventHandler) error { return nil }

type mockLogger struct{}

func (m *mockLogger) Debug(args ...any)                            {}
func (m *mockLogger) Debugf(template string, args ...any)          {}
func (m *mockLogger) Debugw(msg string, keysAndValues ...any)      {}
func (m *mockLogger) Info(args ...any)                             {}
func (m *mockLogger) Infof(template string, args ...any)           {}
func (m *mockLogger) Infow(msg string, keysAndValues ...any)       {}
func (m *mockLogger) Warn(args ...any)                             {}
func (m *mockLogger) Warnf(template string, args ...any)           {}
func (m *mockLogger) Warnw(msg string, keysAndValues ...any)       {}
func (m *mockLogger) Error(args ...any)                            {}
func (m *mockLogger) Errorf(template string, args ...any)          {}
func (m *mockLogger) Errorw(msg string, keysAndValues ...any)      {}
func (m *mockLogger) Fatal(args ...any)                            {}
func (m *mockLogger) Fatalf(template string, args ...any)          {}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any)      {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any) {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any) {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any) {}

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	readRepo := &mockReadRepo{
		views: []*announcerepo.AnnouncementView{
			{ID: announceentity.NewAnnouncementID(), TitleEn: "A1", Priority: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
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
	t.Parallel()

	id := announceentity.NewAnnouncementID()
	readRepo := &mockReadRepo{
		view: &announcerepo.AnnouncementView{
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
	t.Parallel()

	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/announcements/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Update_Success(t *testing.T) {
	t.Parallel()

	a, _ := announceentity.NewAnnouncement(
		shared.Lang{Uz: "t", Ru: "t", En: "t"},
		shared.Lang{Uz: "c", Ru: "c", En: "c"},
		1, nil, nil,
	)
	repo := &mockRepo{
		findFn: func(_ context.Context, id announceentity.AnnouncementID) (*announceentity.Announcement, error) {
			if id == a.TypedID() {
				return a, nil
			}
			return nil, announceentity.ErrAnnouncementNotFound
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
	t.Parallel()

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
	t.Parallel()

	repo := &mockRepo{}
	router := setupRouter(repo, &mockReadRepo{})

	id := announceentity.NewAnnouncementID()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/announcements/"+id.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	t.Parallel()

	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/announcements/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- Additional error-path, parsing, and pagination tests ---

func TestHandler_Get_NotFound(t *testing.T) {
	// mockReadRepo with no view set returns ErrAnnouncementNotFound
	readRepo := &mockReadRepo{}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/announcements/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("expected non-200 for not-found announcement, got %d", w.Code)
	}
}

func TestHandler_List_DefaultPagination(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*announcerepo.AnnouncementView{},
		total: 0,
	}
	router := setupRouter(&mockRepo{}, readRepo)

	// No query params — should use default pagination and succeed
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/announcements", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for default pagination, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_List_InvalidLimit(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/announcements?limit=abc", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid limit, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Create_EmptyBody(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/announcements", bytes.NewBufferString(``))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty body, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/announcements", bytes.NewBufferString(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Delete_InvalidUUID(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/announcements/not-a-uuid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_InvalidUUID(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/announcements/not-a-uuid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Update_InvalidUUID(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/announcements/not-a-uuid", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d: %s", w.Code, w.Body.String())
	}
}
