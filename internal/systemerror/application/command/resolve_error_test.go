package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/shared/application"
	shared "gct/internal/shared/domain"
	"gct/internal/systemerror/domain"

	"github.com/google/uuid"
)

// --- Mocks ---

type mockSystemErrorRepo struct {
	saved   *domain.SystemError
	updated *domain.SystemError
	findFn  func(ctx context.Context, id uuid.UUID) (*domain.SystemError, error)
}

func (m *mockSystemErrorRepo) Save(_ context.Context, e *domain.SystemError) error {
	m.saved = e
	return nil
}

func (m *mockSystemErrorRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.SystemError, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrSystemErrorNotFound
}

func (m *mockSystemErrorRepo) Update(_ context.Context, e *domain.SystemError) error {
	m.updated = e
	return nil
}

func (m *mockSystemErrorRepo) List(_ context.Context, _ domain.SystemErrorFilter) ([]*domain.SystemError, int64, error) {
	return nil, 0, nil
}

type mockEventBus struct {
	published []shared.DomainEvent
}

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

// --- Tests: CreateSystemError ---

func TestCreateSystemErrorHandler_Handle(t *testing.T) {
	repo := &mockSystemErrorRepo{}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateSystemErrorHandler(repo, eb, log)

	stack := "goroutine 1 [running]:\nmain.main()"
	svc := "api-gateway"
	reqID := uuid.New()
	ip := "10.0.0.1"
	path := "/api/v1/users"
	method := "POST"

	cmd := CreateSystemErrorCommand{
		Code:        "ERR_500",
		Message:     "internal server error",
		StackTrace:  &stack,
		Metadata:    map[string]any{"key": "val"},
		Severity:    "critical",
		ServiceName: &svc,
		RequestID:   &reqID,
		IPAddress:   &ip,
		Path:        &path,
		Method:      &method,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.saved == nil {
		t.Fatal("expected system error to be saved")
	}
	if repo.saved.Code() != "ERR_500" {
		t.Errorf("expected code ERR_500, got %s", repo.saved.Code())
	}
	if repo.saved.Severity() != "critical" {
		t.Errorf("expected severity critical, got %s", repo.saved.Severity())
	}
	if repo.saved.StackTrace() == nil || *repo.saved.StackTrace() != stack {
		t.Error("stack trace not set")
	}
	if repo.saved.ServiceName() == nil || *repo.saved.ServiceName() != "api-gateway" {
		t.Error("service name not set")
	}
	if len(eb.published) == 0 {
		t.Fatal("expected events to be published")
	}
	if eb.published[0].EventName() != "system_error.recorded" {
		t.Errorf("expected system_error.recorded, got %s", eb.published[0].EventName())
	}
}

func TestCreateSystemErrorHandler_MinimalFields(t *testing.T) {
	repo := &mockSystemErrorRepo{}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateSystemErrorHandler(repo, eb, log)

	err := handler.Handle(context.Background(), CreateSystemErrorCommand{
		Code:     "ERR_400",
		Message:  "bad request",
		Severity: "warning",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.saved == nil {
		t.Fatal("expected system error to be saved")
	}
	if repo.saved.StackTrace() != nil {
		t.Error("stack trace should be nil")
	}
	if repo.saved.ServiceName() != nil {
		t.Error("service name should be nil")
	}
}

// --- Tests: ResolveError ---

func TestResolveErrorHandler_Handle(t *testing.T) {
	se := domain.NewSystemError("ERR_500", "test error", "critical")

	repo := &mockSystemErrorRepo{
		findFn: func(_ context.Context, id uuid.UUID) (*domain.SystemError, error) {
			if id == se.ID() {
				return se, nil
			}
			return nil, domain.ErrSystemErrorNotFound
		},
	}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewResolveErrorHandler(repo, eb, log)

	resolverID := uuid.New()
	err := handler.Handle(context.Background(), ResolveErrorCommand{
		ID:         se.ID(),
		ResolvedBy: resolverID,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.updated == nil {
		t.Fatal("expected system error to be updated")
	}
	if !repo.updated.IsResolved() {
		t.Error("expected system error to be resolved")
	}
	if repo.updated.ResolvedBy() == nil || *repo.updated.ResolvedBy() != resolverID {
		t.Error("resolvedBy not set correctly")
	}
	if repo.updated.ResolvedAt() == nil {
		t.Error("resolvedAt should be set")
	}

	found := false
	for _, e := range eb.published {
		if e.EventName() == "system_error.resolved" {
			found = true
		}
	}
	if !found {
		t.Error("expected system_error.resolved event")
	}
}

func TestResolveErrorHandler_NotFound(t *testing.T) {
	repo := &mockSystemErrorRepo{}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewResolveErrorHandler(repo, eb, log)

	err := handler.Handle(context.Background(), ResolveErrorCommand{
		ID:         uuid.New(),
		ResolvedBy: uuid.New(),
	})
	if err == nil {
		t.Fatal("expected error for non-existent system error")
	}
}

func TestResolveErrorHandler_AlreadyResolved(t *testing.T) {
	resolverID := uuid.New()
	now := time.Now()
	se := domain.ReconstructSystemError(
		uuid.New(), time.Now(),
		"ERR_500", "test", nil, nil, "critical",
		nil, nil, nil, nil, nil, nil,
		true, &now, &resolverID,
	)

	repo := &mockSystemErrorRepo{
		findFn: func(_ context.Context, id uuid.UUID) (*domain.SystemError, error) {
			return se, nil
		},
	}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewResolveErrorHandler(repo, eb, log)

	// Should be idempotent
	err := handler.Handle(context.Background(), ResolveErrorCommand{
		ID:         se.ID(),
		ResolvedBy: uuid.New(),
	})
	if err != nil {
		t.Fatalf("resolve on already-resolved should be idempotent, got: %v", err)
	}
}

// --- Error paths ---

var errRepoSave = errors.New("repo save failed")
var errRepoUpdate = errors.New("repo update failed")

func TestCreateSystemErrorHandler_RepoError(t *testing.T) {
	repo := &mockSystemErrorRepo{}
	repo2 := &errorSystemErrorRepo{saveErr: errRepoSave}
	eb := &mockEventBus{}
	log := &mockLogger{}

	_ = repo // unused, just to keep the mock available

	handler := NewCreateSystemErrorHandler(repo2, eb, log)
	err := handler.Handle(context.Background(), CreateSystemErrorCommand{
		Code: "ERR", Message: "fail", Severity: "low",
	})
	if !errors.Is(err, errRepoSave) {
		t.Fatalf("expected errRepoSave, got: %v", err)
	}
}

func TestResolveErrorHandler_RepoUpdateError(t *testing.T) {
	se := domain.NewSystemError("ERR", "test", "low")

	repo := &errorSystemErrorRepo{
		findFn:    func(_ context.Context, _ uuid.UUID) (*domain.SystemError, error) { return se, nil },
		updateErr: errRepoUpdate,
	}
	eb := &mockEventBus{}
	log := &mockLogger{}

	handler := NewResolveErrorHandler(repo, eb, log)
	err := handler.Handle(context.Background(), ResolveErrorCommand{
		ID: se.ID(), ResolvedBy: uuid.New(),
	})
	if !errors.Is(err, errRepoUpdate) {
		t.Fatalf("expected errRepoUpdate, got: %v", err)
	}
}

type errorSystemErrorRepo struct {
	saveErr   error
	updateErr error
	findFn    func(ctx context.Context, id uuid.UUID) (*domain.SystemError, error)
}

func (m *errorSystemErrorRepo) Save(_ context.Context, _ *domain.SystemError) error {
	return m.saveErr
}

func (m *errorSystemErrorRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.SystemError, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrSystemErrorNotFound
}

func (m *errorSystemErrorRepo) Update(_ context.Context, _ *domain.SystemError) error {
	return m.updateErr
}

func (m *errorSystemErrorRepo) List(_ context.Context, _ domain.SystemErrorFilter) ([]*domain.SystemError, int64, error) {
	return nil, 0, nil
}
