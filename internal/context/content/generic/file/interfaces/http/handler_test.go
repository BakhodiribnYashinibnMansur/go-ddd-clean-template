package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/context/content/generic/file"
	"gct/internal/context/content/generic/file/application/command"
	"gct/internal/context/content/generic/file/application/query"
	fileentity "gct/internal/context/content/generic/file/domain/entity"
	filerepo "gct/internal/context/content/generic/file/domain/repository"
	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- Mocks ---

type mockRepo struct {
	saved *fileentity.File
}

func (m *mockRepo) Save(_ context.Context, f *fileentity.File) error {
	m.saved = f
	return nil
}

type mockReadRepo struct {
	view  *filerepo.FileView
	views []*filerepo.FileView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id fileentity.FileID) (*filerepo.FileView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, fileentity.ErrFileNotFound
}
func (m *mockReadRepo) List(_ context.Context, _ filerepo.FileFilter) ([]*filerepo.FileView, int64, error) {
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

	bc := &file.BoundedContext{
		CreateFile: command.NewCreateFileHandler(repo, eb, l),
		GetFile:    query.NewGetFileHandler(readRepo, l),
		ListFiles:  query.NewListFilesHandler(readRepo, l),
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
		Name:         "avatar.png",
		OriginalName: "my-avatar.png",
		MimeType:     "image/png",
		Size:         1024,
		Path:         "/uploads/avatar.png",
		URL:          "https://cdn.example.com/avatar.png",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/files", bytes.NewBuffer(jsonBody))
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
	req, _ := http.NewRequest("POST", "/api/v1/files", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_List_Success(t *testing.T) {
	t.Parallel()

	readRepo := &mockReadRepo{
		views: []*filerepo.FileView{
			{ID: fileentity.NewFileID(), Name: "file1.png", MimeType: "image/png", Size: 100, CreatedAt: time.Now()},
		},
		total: 1,
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/files?limit=10&offset=0", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_Success(t *testing.T) {
	t.Parallel()

	id := fileentity.NewFileID()
	readRepo := &mockReadRepo{
		view: &filerepo.FileView{ID: id, Name: "doc.pdf", MimeType: "application/pdf", Size: 2048, CreatedAt: time.Now()},
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/files/"+id.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_InvalidID(t *testing.T) {
	t.Parallel()

	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/files/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Get_NotFound(t *testing.T) {
	// readRepo has no view set, so FindByID returns fileentity.ErrFileNotFound
	readRepo := &mockReadRepo{}
	router := setupRouter(&mockRepo{}, readRepo)

	id := uuid.New()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/files/"+id.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("expected non-200 for missing file, got %d", w.Code)
	}
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/files", bytes.NewBufferString(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_List_DefaultPagination(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*filerepo.FileView{},
		total: 0,
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	// No query params — should use default pagination and return 200
	req, _ := http.NewRequest("GET", "/api/v1/files", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
