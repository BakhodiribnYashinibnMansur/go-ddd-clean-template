package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/context/admin/supporting/dataexport"
	"gct/internal/context/admin/supporting/dataexport/application/command"
	"gct/internal/context/admin/supporting/dataexport/application/query"
	exportentity "gct/internal/context/admin/supporting/dataexport/domain/entity"
	exportrepo "gct/internal/context/admin/supporting/dataexport/domain/repository"
	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"

	"gct/internal/kernel/outbox"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- Mocks ---

type mockRepo struct {
	saved   *exportentity.DataExport
	updated *exportentity.DataExport
	deleted exportentity.DataExportID
	findFn  func(ctx context.Context, id exportentity.DataExportID) (*exportentity.DataExport, error)
}

func (m *mockRepo) Save(_ context.Context, e *exportentity.DataExport) error {
	m.saved = e
	return nil
}
func (m *mockRepo) FindByID(ctx context.Context, id exportentity.DataExportID) (*exportentity.DataExport, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, exportentity.ErrDataExportNotFound
}
func (m *mockRepo) Update(_ context.Context, e *exportentity.DataExport) error {
	m.updated = e
	return nil
}
func (m *mockRepo) Delete(_ context.Context, id exportentity.DataExportID) error {
	m.deleted = id
	return nil
}

type mockReadRepo struct {
	view  *exportrepo.DataExportView
	views []*exportrepo.DataExportView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id exportentity.DataExportID) (*exportrepo.DataExportView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, exportentity.ErrDataExportNotFound
}
func (m *mockReadRepo) List(_ context.Context, _ exportrepo.DataExportFilter) ([]*exportrepo.DataExportView, int64, error) {
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

	bc := &dataexport.BoundedContext{
		CreateDataExport: command.NewCreateDataExportHandler(repo, outbox.NewEventCommitter(nil, nil, eb, l), l),
		UpdateDataExport: command.NewUpdateDataExportHandler(repo, outbox.NewEventCommitter(nil, nil, eb, l), l),
		DeleteDataExport: command.NewDeleteDataExportHandler(repo, l),
		GetDataExport:    query.NewGetDataExportHandler(readRepo, l),
		ListDataExports:  query.NewListDataExportsHandler(readRepo, l),
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
		UserID:   uuid.New(),
		DataType: "users",
		Format:   "csv",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/data-exports", bytes.NewBuffer(jsonBody))
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
	req, _ := http.NewRequest("POST", "/api/v1/data-exports", bytes.NewBufferString(`not json`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_List_Success(t *testing.T) {
	t.Parallel()

	readRepo := &mockReadRepo{
		views: []*exportrepo.DataExportView{
			{ID: exportentity.NewDataExportID(), UserID: uuid.New(), DataType: "users", Format: "csv", Status: "PENDING", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		total: 1,
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/data-exports?limit=10&offset=0", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_Success(t *testing.T) {
	t.Parallel()

	id := exportentity.NewDataExportID()
	readRepo := &mockReadRepo{
		view: &exportrepo.DataExportView{
			ID: id, UserID: uuid.New(), DataType: "users", Format: "csv", Status: "COMPLETED", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/data-exports/"+id.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_InvalidID(t *testing.T) {
	t.Parallel()

	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/data-exports/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Delete_Success(t *testing.T) {
	t.Parallel()

	repo := &mockRepo{}
	router := setupRouter(repo, &mockReadRepo{})

	id := exportentity.NewDataExportID()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/data-exports/"+id.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	t.Parallel()

	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/data-exports/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Get_NotFound(t *testing.T) {
	// readRepo has no view set, so FindByID returns ErrDataExportNotFound
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	id := exportentity.NewDataExportID()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/data-exports/"+id.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("expected non-200 for missing data export, got %d", w.Code)
	}
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/data-exports", bytes.NewBufferString(`not json at all`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestHandler_List_DefaultPagination(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*exportrepo.DataExportView{},
		total: 0,
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/data-exports", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with default pagination, got %d: %s", w.Code, w.Body.String())
	}
}
