package command

import (
	"context"
	"testing"

	"gct/internal/context/iam/supporting/audit/domain"
	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// --- Mock Repository ---

type mockAuditLogRepository struct {
	savedAuditLog *domain.AuditLog
	saveErr       error
}

func (m *mockAuditLogRepository) Save(ctx context.Context, auditLog *domain.AuditLog) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedAuditLog = auditLog
	return nil
}

// --- Mock EventBus ---

type mockEventBus struct {
	publishedEvents []shared.DomainEvent
}

func (m *mockEventBus) Publish(ctx context.Context, events ...shared.DomainEvent) error {
	m.publishedEvents = append(m.publishedEvents, events...)
	return nil
}

func (m *mockEventBus) Subscribe(eventName string, handler application.EventHandler) error {
	return nil
}

// --- Mock Logger ---

type mockLogger struct{}

func (m *mockLogger) Debug(args ...any)                                            {}
func (m *mockLogger) Debugf(template string, args ...any)                          {}
func (m *mockLogger) Debugw(msg string, keysAndValues ...any)                      {}
func (m *mockLogger) Info(args ...any)                                             {}
func (m *mockLogger) Infof(template string, args ...any)                           {}
func (m *mockLogger) Infow(msg string, keysAndValues ...any)                       {}
func (m *mockLogger) Warn(args ...any)                                             {}
func (m *mockLogger) Warnf(template string, args ...any)                           {}
func (m *mockLogger) Warnw(msg string, keysAndValues ...any)                       {}
func (m *mockLogger) Error(args ...any)                                            {}
func (m *mockLogger) Errorf(template string, args ...any)                          {}
func (m *mockLogger) Errorw(msg string, keysAndValues ...any)                      {}
func (m *mockLogger) Fatal(args ...any)                                            {}
func (m *mockLogger) Fatalf(template string, args ...any)                          {}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any)                      {}
func (m *mockLogger) Debugc(ctx context.Context, msg string, keysAndValues ...any) {}
func (m *mockLogger) Infoc(ctx context.Context, msg string, keysAndValues ...any)  {}
func (m *mockLogger) Warnc(ctx context.Context, msg string, keysAndValues ...any)  {}
func (m *mockLogger) Errorc(ctx context.Context, msg string, keysAndValues ...any) {}
func (m *mockLogger) Fatalc(ctx context.Context, msg string, keysAndValues ...any) {}

// --- Tests ---

func TestCreateAuditLogHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockAuditLogRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateAuditLogHandler(repo, eventBus, log)

	userID := uuid.New()
	ip := "10.0.0.1"

	cmd := CreateAuditLogCommand{
		UserID:    &userID,
		Action:    domain.AuditActionLogin,
		IPAddress: &ip,
		Success:   true,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.savedAuditLog == nil {
		t.Fatal("expected audit log to be saved, but it was nil")
	}

	if repo.savedAuditLog.Action() != domain.AuditActionLogin {
		t.Errorf("expected action LOGIN, got %s", repo.savedAuditLog.Action())
	}

	if *repo.savedAuditLog.UserID() != userID {
		t.Error("expected userID to match")
	}

	if *repo.savedAuditLog.IPAddress() != "10.0.0.1" {
		t.Error("expected ipAddress to be 10.0.0.1")
	}

	if !repo.savedAuditLog.Success() {
		t.Error("expected success to be true")
	}

	if len(eventBus.publishedEvents) == 0 {
		t.Fatal("expected at least one event to be published")
	}

	if eventBus.publishedEvents[0].EventName() != "audit_log.created" {
		t.Errorf("expected event audit_log.created, got %s", eventBus.publishedEvents[0].EventName())
	}
}

func TestCreateAuditLogHandler_RepoError(t *testing.T) {
	t.Parallel()

	repoErr := shared.NewDomainError("TEST_ERROR", "test error")
	repo := &mockAuditLogRepository{saveErr: repoErr}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateAuditLogHandler(repo, eventBus, log)

	cmd := CreateAuditLogCommand{
		Action:  domain.AuditActionLogout,
		Success: true,
	}

	err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}

	if repo.savedAuditLog != nil {
		t.Error("expected no audit log to be saved on error")
	}

	if len(eventBus.publishedEvents) != 0 {
		t.Error("expected no events to be published on repo error")
	}
}

func TestCreateAuditLogHandler_WithMetadata(t *testing.T) {
	t.Parallel()

	repo := &mockAuditLogRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewCreateAuditLogHandler(repo, eventBus, log)

	meta := map[string]string{"browser": "Chrome", "version": "120"}

	cmd := CreateAuditLogCommand{
		Action:   domain.AuditActionUserUpdate,
		Success:  true,
		Metadata: meta,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	if repo.savedAuditLog.Metadata()["browser"] != "Chrome" {
		t.Error("expected metadata browser to be Chrome")
	}
}
