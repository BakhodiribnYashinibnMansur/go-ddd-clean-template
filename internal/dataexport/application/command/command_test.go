package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/dataexport/application/command"
	"gct/internal/dataexport/domain"
	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Mock infrastructure
// ---------------------------------------------------------------------------

type mockWriteRepo struct {
	savedEntity   *domain.DataExport
	updatedEntity *domain.DataExport
	deletedID     uuid.UUID
	findByIDFn    func(ctx context.Context, id uuid.UUID) (*domain.DataExport, error)
	saveFn        func(ctx context.Context, entity *domain.DataExport) error
	updateFn      func(ctx context.Context, entity *domain.DataExport) error
	deleteFn      func(ctx context.Context, id uuid.UUID) error
}

func (m *mockWriteRepo) Save(ctx context.Context, entity *domain.DataExport) error {
	m.savedEntity = entity
	if m.saveFn != nil {
		return m.saveFn(ctx, entity)
	}
	return nil
}

func (m *mockWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.DataExport, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, domain.ErrDataExportNotFound
}

func (m *mockWriteRepo) Update(ctx context.Context, entity *domain.DataExport) error {
	m.updatedEntity = entity
	if m.updateFn != nil {
		return m.updateFn(ctx, entity)
	}
	return nil
}

func (m *mockWriteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	m.deletedID = id
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

type mockEventBus struct {
	publishedEvents []shared.DomainEvent
	publishFn       func(ctx context.Context, events ...shared.DomainEvent) error
}

func (m *mockEventBus) Publish(ctx context.Context, events ...shared.DomainEvent) error {
	m.publishedEvents = append(m.publishedEvents, events...)
	if m.publishFn != nil {
		return m.publishFn(ctx, events...)
	}
	return nil
}

func (m *mockEventBus) Subscribe(_ string, _ func(context.Context, shared.DomainEvent) error) error {
	return nil
}

type mockLogger struct{}

func (m *mockLogger) Debug(args ...any)                                {}
func (m *mockLogger) Debugf(template string, args ...any)              {}
func (m *mockLogger) Debugw(msg string, keysAndValues ...any)          {}
func (m *mockLogger) Info(args ...any)                                 {}
func (m *mockLogger) Infof(template string, args ...any)               {}
func (m *mockLogger) Infow(msg string, keysAndValues ...any)           {}
func (m *mockLogger) Warn(args ...any)                                 {}
func (m *mockLogger) Warnf(template string, args ...any)               {}
func (m *mockLogger) Warnw(msg string, keysAndValues ...any)           {}
func (m *mockLogger) Error(args ...any)                                {}
func (m *mockLogger) Errorf(template string, args ...any)              {}
func (m *mockLogger) Errorw(msg string, keysAndValues ...any)          {}
func (m *mockLogger) Fatal(args ...any)                                {}
func (m *mockLogger) Fatalf(template string, args ...any)              {}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any)          {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any)     {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)      {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)      {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any)     {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any)     {}

// ---------------------------------------------------------------------------
// Tests: CreateDataExportHandler
// ---------------------------------------------------------------------------

func TestCreateDataExportHandler_Success(t *testing.T) {
	repo := &mockWriteRepo{}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewCreateDataExportHandler(repo, eb, l)

	cmd := command.CreateDataExportCommand{
		UserID:   uuid.New(),
		DataType: "users",
		Format:   "csv",
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repo.savedEntity == nil {
		t.Fatal("expected entity to be saved")
	}
	if repo.savedEntity.UserID() != cmd.UserID {
		t.Fatalf("expected user ID %s, got %s", cmd.UserID, repo.savedEntity.UserID())
	}
	if repo.savedEntity.DataType() != "users" {
		t.Fatalf("expected data type users, got %s", repo.savedEntity.DataType())
	}
	if repo.savedEntity.Format() != "csv" {
		t.Fatalf("expected format csv, got %s", repo.savedEntity.Format())
	}
	if repo.savedEntity.Status() != domain.ExportStatusPending {
		t.Fatalf("expected status PENDING, got %s", repo.savedEntity.Status())
	}
	if len(eb.publishedEvents) != 1 {
		t.Fatalf("expected 1 event published, got %d", len(eb.publishedEvents))
	}
	if eb.publishedEvents[0].EventName() != "dataexport.requested" {
		t.Fatalf("expected dataexport.requested event, got %s", eb.publishedEvents[0].EventName())
	}
}

func TestCreateDataExportHandler_RepoSaveError(t *testing.T) {
	repoErr := errors.New("db connection failed")
	repo := &mockWriteRepo{
		saveFn: func(_ context.Context, _ *domain.DataExport) error {
			return repoErr
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewCreateDataExportHandler(repo, eb, l)

	cmd := command.CreateDataExportCommand{
		UserID:   uuid.New(),
		DataType: "orders",
		Format:   "xlsx",
	}

	err := h.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
	if len(eb.publishedEvents) != 0 {
		t.Fatalf("expected 0 events published on save error, got %d", len(eb.publishedEvents))
	}
}

func TestCreateDataExportHandler_EventBusError(t *testing.T) {
	repo := &mockWriteRepo{}
	eb := &mockEventBus{
		publishFn: func(_ context.Context, _ ...shared.DomainEvent) error {
			return errors.New("event bus down")
		},
	}
	l := &mockLogger{}
	h := command.NewCreateDataExportHandler(repo, eb, l)

	cmd := command.CreateDataExportCommand{
		UserID:   uuid.New(),
		DataType: "logs",
		Format:   "json",
	}

	// Event bus failure is non-fatal — handler returns nil
	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error (event bus failure is non-fatal), got %v", err)
	}
	if repo.savedEntity == nil {
		t.Fatal("entity should still be saved even if event bus fails")
	}
}

// ---------------------------------------------------------------------------
// Tests: UpdateDataExportHandler
// ---------------------------------------------------------------------------

func TestUpdateDataExportHandler_StatusProcessing(t *testing.T) {
	exportID := uuid.New()
	existing := domain.ReconstructDataExport(
		exportID, time.Now(), time.Now(),
		uuid.New(), "users", "csv", domain.ExportStatusPending, nil, nil,
	)

	repo := &mockWriteRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.DataExport, error) {
			if id == exportID {
				return existing, nil
			}
			return nil, domain.ErrDataExportNotFound
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	status := domain.ExportStatusProcessing
	cmd := command.UpdateDataExportCommand{
		ID:     exportID,
		Status: &status,
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedEntity == nil {
		t.Fatal("expected entity to be updated")
	}
	if repo.updatedEntity.Status() != domain.ExportStatusProcessing {
		t.Fatalf("expected status PROCESSING, got %s", repo.updatedEntity.Status())
	}
}

func TestUpdateDataExportHandler_StatusCompleted(t *testing.T) {
	exportID := uuid.New()
	existing := domain.ReconstructDataExport(
		exportID, time.Now(), time.Now(),
		uuid.New(), "users", "csv", domain.ExportStatusProcessing, nil, nil,
	)

	repo := &mockWriteRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.DataExport, error) {
			if id == exportID {
				return existing, nil
			}
			return nil, domain.ErrDataExportNotFound
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	status := domain.ExportStatusCompleted
	fileURL := "https://example.com/export.csv"
	cmd := command.UpdateDataExportCommand{
		ID:      exportID,
		Status:  &status,
		FileURL: &fileURL,
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedEntity.Status() != domain.ExportStatusCompleted {
		t.Fatalf("expected status COMPLETED, got %s", repo.updatedEntity.Status())
	}
	if repo.updatedEntity.FileURL() == nil || *repo.updatedEntity.FileURL() != fileURL {
		t.Fatalf("expected file URL %s, got %v", fileURL, repo.updatedEntity.FileURL())
	}
	// Complete raises an ExportCompleted event
	if len(eb.publishedEvents) == 0 {
		t.Fatal("expected at least one event published")
	}
}

func TestUpdateDataExportHandler_StatusFailed(t *testing.T) {
	exportID := uuid.New()
	existing := domain.ReconstructDataExport(
		exportID, time.Now(), time.Now(),
		uuid.New(), "orders", "xlsx", domain.ExportStatusProcessing, nil, nil,
	)

	repo := &mockWriteRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.DataExport, error) {
			if id == exportID {
				return existing, nil
			}
			return nil, domain.ErrDataExportNotFound
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	status := domain.ExportStatusFailed
	errMsg := "disk full"
	cmd := command.UpdateDataExportCommand{
		ID:     exportID,
		Status: &status,
		Error:  &errMsg,
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedEntity.Status() != domain.ExportStatusFailed {
		t.Fatalf("expected status FAILED, got %s", repo.updatedEntity.Status())
	}
	if repo.updatedEntity.Error() == nil || *repo.updatedEntity.Error() != errMsg {
		t.Fatalf("expected error msg %q, got %v", errMsg, repo.updatedEntity.Error())
	}
}

func TestUpdateDataExportHandler_NotFound(t *testing.T) {
	repo := &mockWriteRepo{} // default findByIDFn returns ErrDataExportNotFound
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	status := domain.ExportStatusProcessing
	cmd := command.UpdateDataExportCommand{
		ID:     uuid.New(),
		Status: &status,
	}

	err := h.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrDataExportNotFound) {
		t.Fatalf("expected ErrDataExportNotFound, got %v", err)
	}
}

func TestUpdateDataExportHandler_RepoUpdateError(t *testing.T) {
	exportID := uuid.New()
	existing := domain.ReconstructDataExport(
		exportID, time.Now(), time.Now(),
		uuid.New(), "users", "csv", domain.ExportStatusPending, nil, nil,
	)
	repoErr := errors.New("update failed")

	repo := &mockWriteRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.DataExport, error) {
			if id == exportID {
				return existing, nil
			}
			return nil, domain.ErrDataExportNotFound
		},
		updateFn: func(_ context.Context, _ *domain.DataExport) error {
			return repoErr
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	status := domain.ExportStatusProcessing
	cmd := command.UpdateDataExportCommand{
		ID:     exportID,
		Status: &status,
	}

	err := h.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestUpdateDataExportHandler_NilStatus(t *testing.T) {
	exportID := uuid.New()
	existing := domain.ReconstructDataExport(
		exportID, time.Now(), time.Now(),
		uuid.New(), "users", "csv", domain.ExportStatusPending, nil, nil,
	)

	repo := &mockWriteRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.DataExport, error) {
			if id == exportID {
				return existing, nil
			}
			return nil, domain.ErrDataExportNotFound
		},
	}
	eb := &mockEventBus{}
	l := &mockLogger{}
	h := command.NewUpdateDataExportHandler(repo, eb, l)

	// No status change — should still update (no-op transition)
	cmd := command.UpdateDataExportCommand{ID: exportID}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedEntity.Status() != domain.ExportStatusPending {
		t.Fatalf("expected status unchanged (PENDING), got %s", repo.updatedEntity.Status())
	}
}

// ---------------------------------------------------------------------------
// Tests: DeleteDataExportHandler
// ---------------------------------------------------------------------------

func TestDeleteDataExportHandler_Success(t *testing.T) {
	repo := &mockWriteRepo{}
	l := &mockLogger{}
	h := command.NewDeleteDataExportHandler(repo, l)

	exportID := uuid.New()
	cmd := command.DeleteDataExportCommand{ID: exportID}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.deletedID != exportID {
		t.Fatalf("expected deleted ID %s, got %s", exportID, repo.deletedID)
	}
}

func TestDeleteDataExportHandler_RepoError(t *testing.T) {
	repoErr := errors.New("delete failed")
	repo := &mockWriteRepo{
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return repoErr
		},
	}
	l := &mockLogger{}
	h := command.NewDeleteDataExportHandler(repo, l)

	cmd := command.DeleteDataExportCommand{ID: uuid.New()}

	err := h.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
