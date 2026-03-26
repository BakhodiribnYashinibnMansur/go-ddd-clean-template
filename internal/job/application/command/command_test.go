package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/job/domain"
	"gct/internal/shared/application"
	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Mock infrastructure
// ---------------------------------------------------------------------------

type mockJobRepo struct {
	savedJob   *domain.Job
	updatedJob *domain.Job
	deletedID  uuid.UUID
	findByIDFn func(ctx context.Context, id uuid.UUID) (*domain.Job, error)
	saveFn     func(ctx context.Context, entity *domain.Job) error
	updateFn   func(ctx context.Context, entity *domain.Job) error
	deleteFn   func(ctx context.Context, id uuid.UUID) error
}

func (m *mockJobRepo) Save(ctx context.Context, entity *domain.Job) error {
	m.savedJob = entity
	if m.saveFn != nil {
		return m.saveFn(ctx, entity)
	}
	return nil
}

func (m *mockJobRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, domain.ErrJobNotFound
}

func (m *mockJobRepo) Update(ctx context.Context, entity *domain.Job) error {
	m.updatedJob = entity
	if m.updateFn != nil {
		return m.updateFn(ctx, entity)
	}
	return nil
}

func (m *mockJobRepo) Delete(ctx context.Context, id uuid.UUID) error {
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

func (m *mockEventBus) Subscribe(_ string, _ application.EventHandler) error { return nil }

type mockLogger struct{}

func (m *mockLogger) Debug(args ...any)                                    {}
func (m *mockLogger) Debugf(template string, args ...any)                  {}
func (m *mockLogger) Debugw(msg string, keysAndValues ...any)              {}
func (m *mockLogger) Info(args ...any)                                     {}
func (m *mockLogger) Infof(template string, args ...any)                   {}
func (m *mockLogger) Infow(msg string, keysAndValues ...any)               {}
func (m *mockLogger) Warn(args ...any)                                     {}
func (m *mockLogger) Warnf(template string, args ...any)                   {}
func (m *mockLogger) Warnw(msg string, keysAndValues ...any)               {}
func (m *mockLogger) Error(args ...any)                                    {}
func (m *mockLogger) Errorf(template string, args ...any)                  {}
func (m *mockLogger) Errorw(msg string, keysAndValues ...any)              {}
func (m *mockLogger) Fatal(args ...any)                                    {}
func (m *mockLogger) Fatalf(template string, args ...any)                  {}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any)              {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any)         {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)          {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)          {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any)         {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any)         {}

// ---------------------------------------------------------------------------
// CreateJobHandler tests
// ---------------------------------------------------------------------------

func TestCreateJobHandler_Success(t *testing.T) {
	repo := &mockJobRepo{}
	eb := &mockEventBus{}
	h := NewCreateJobHandler(repo, eb, &mockLogger{})

	scheduledAt := time.Now().Add(time.Hour)
	cmd := CreateJobCommand{
		TaskName:    "send_email",
		Payload:     map[string]any{"to": "user@example.com"},
		MaxAttempts: 3,
		ScheduledAt: &scheduledAt,
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repo.savedJob == nil {
		t.Fatal("expected job to be saved")
	}
	if repo.savedJob.TaskName() != "send_email" {
		t.Fatalf("expected task name send_email, got %s", repo.savedJob.TaskName())
	}
	if repo.savedJob.Status() != domain.JobStatusPending {
		t.Fatalf("expected status PENDING, got %s", repo.savedJob.Status())
	}
	if repo.savedJob.MaxAttempts() != 3 {
		t.Fatalf("expected max attempts 3, got %d", repo.savedJob.MaxAttempts())
	}
	if len(eb.publishedEvents) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(eb.publishedEvents))
	}
	if eb.publishedEvents[0].EventName() != "job.scheduled" {
		t.Fatalf("expected job.scheduled event, got %s", eb.publishedEvents[0].EventName())
	}
}

func TestCreateJobHandler_NilPayload(t *testing.T) {
	repo := &mockJobRepo{}
	eb := &mockEventBus{}
	h := NewCreateJobHandler(repo, eb, &mockLogger{})

	cmd := CreateJobCommand{
		TaskName:    "cleanup",
		Payload:     nil,
		MaxAttempts: 1,
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.savedJob == nil {
		t.Fatal("expected job to be saved")
	}
	if repo.savedJob.Payload() == nil {
		t.Fatal("expected payload to be initialized (not nil)")
	}
}

func TestCreateJobHandler_RepoSaveError(t *testing.T) {
	repoErr := errors.New("db connection failed")
	repo := &mockJobRepo{
		saveFn: func(_ context.Context, _ *domain.Job) error {
			return repoErr
		},
	}
	eb := &mockEventBus{}
	h := NewCreateJobHandler(repo, eb, &mockLogger{})

	cmd := CreateJobCommand{
		TaskName:    "process",
		MaxAttempts: 1,
	}

	err := h.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
	if len(eb.publishedEvents) != 0 {
		t.Fatal("expected no events published when save fails")
	}
}

func TestCreateJobHandler_EventBusError_DoesNotFail(t *testing.T) {
	repo := &mockJobRepo{}
	eb := &mockEventBus{
		publishFn: func(_ context.Context, _ ...shared.DomainEvent) error {
			return errors.New("event bus down")
		},
	}
	h := NewCreateJobHandler(repo, eb, &mockLogger{})

	cmd := CreateJobCommand{
		TaskName:    "notify",
		MaxAttempts: 1,
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected nil error (event bus failure is non-fatal), got %v", err)
	}
	if repo.savedJob == nil {
		t.Fatal("expected job to be saved despite event bus failure")
	}
}

// ---------------------------------------------------------------------------
// UpdateJobHandler tests
// ---------------------------------------------------------------------------

func TestUpdateJobHandler_TransitionToRunning(t *testing.T) {
	existingJob := domain.NewJob("process_data", nil, 3, nil)
	repo := &mockJobRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Job, error) {
			if id == existingJob.ID() {
				return existingJob, nil
			}
			return nil, domain.ErrJobNotFound
		},
	}
	eb := &mockEventBus{}
	h := NewUpdateJobHandler(repo, eb, &mockLogger{})

	status := domain.JobStatusRunning
	cmd := UpdateJobCommand{
		ID:     existingJob.ID(),
		Status: &status,
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedJob == nil {
		t.Fatal("expected job to be updated")
	}
	if repo.updatedJob.Status() != domain.JobStatusRunning {
		t.Fatalf("expected RUNNING status, got %s", repo.updatedJob.Status())
	}
	if repo.updatedJob.Attempts() != 1 {
		t.Fatalf("expected 1 attempt after start, got %d", repo.updatedJob.Attempts())
	}
}

func TestUpdateJobHandler_TransitionToCompleted(t *testing.T) {
	existingJob := domain.NewJob("process_data", nil, 3, nil)
	existingJob.Start()

	repo := &mockJobRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Job, error) {
			if id == existingJob.ID() {
				return existingJob, nil
			}
			return nil, domain.ErrJobNotFound
		},
	}
	eb := &mockEventBus{}
	h := NewUpdateJobHandler(repo, eb, &mockLogger{})

	status := domain.JobStatusCompleted
	result := map[string]any{"rows_processed": 42}
	cmd := UpdateJobCommand{
		ID:     existingJob.ID(),
		Status: &status,
		Result: result,
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedJob.Status() != domain.JobStatusCompleted {
		t.Fatalf("expected COMPLETED status, got %s", repo.updatedJob.Status())
	}
	if repo.updatedJob.CompletedAt() == nil {
		t.Fatal("expected completed_at to be set")
	}
}

func TestUpdateJobHandler_TransitionToFailed(t *testing.T) {
	existingJob := domain.NewJob("import_data", nil, 3, nil)
	existingJob.Start()

	repo := &mockJobRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Job, error) {
			if id == existingJob.ID() {
				return existingJob, nil
			}
			return nil, domain.ErrJobNotFound
		},
	}
	eb := &mockEventBus{}
	h := NewUpdateJobHandler(repo, eb, &mockLogger{})

	status := domain.JobStatusFailed
	errMsg := "connection timeout"
	cmd := UpdateJobCommand{
		ID:     existingJob.ID(),
		Status: &status,
		Error:  &errMsg,
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedJob.Status() != domain.JobStatusFailed {
		t.Fatalf("expected FAILED status, got %s", repo.updatedJob.Status())
	}
	if repo.updatedJob.Error() == nil || *repo.updatedJob.Error() != "connection timeout" {
		t.Fatal("expected error message to be set")
	}
}

func TestUpdateJobHandler_FailedWithNilError(t *testing.T) {
	existingJob := domain.NewJob("import_data", nil, 3, nil)
	existingJob.Start()

	repo := &mockJobRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Job, error) {
			if id == existingJob.ID() {
				return existingJob, nil
			}
			return nil, domain.ErrJobNotFound
		},
	}
	eb := &mockEventBus{}
	h := NewUpdateJobHandler(repo, eb, &mockLogger{})

	status := domain.JobStatusFailed
	cmd := UpdateJobCommand{
		ID:     existingJob.ID(),
		Status: &status,
		Error:  nil, // no error message provided
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedJob.Error() == nil || *repo.updatedJob.Error() != "" {
		t.Fatalf("expected empty error message, got %v", repo.updatedJob.Error())
	}
}

func TestUpdateJobHandler_JobNotFound(t *testing.T) {
	repo := &mockJobRepo{} // findByIDFn is nil, returns ErrJobNotFound
	eb := &mockEventBus{}
	h := NewUpdateJobHandler(repo, eb, &mockLogger{})

	status := domain.JobStatusRunning
	cmd := UpdateJobCommand{
		ID:     uuid.New(),
		Status: &status,
	}

	err := h.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrJobNotFound) {
		t.Fatalf("expected ErrJobNotFound, got %v", err)
	}
}

func TestUpdateJobHandler_RepoUpdateError(t *testing.T) {
	existingJob := domain.NewJob("task", nil, 1, nil)
	repoErr := errors.New("update failed")
	repo := &mockJobRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Job, error) {
			if id == existingJob.ID() {
				return existingJob, nil
			}
			return nil, domain.ErrJobNotFound
		},
		updateFn: func(_ context.Context, _ *domain.Job) error {
			return repoErr
		},
	}
	eb := &mockEventBus{}
	h := NewUpdateJobHandler(repo, eb, &mockLogger{})

	status := domain.JobStatusRunning
	cmd := UpdateJobCommand{
		ID:     existingJob.ID(),
		Status: &status,
	}

	err := h.Handle(context.Background(), cmd)
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestUpdateJobHandler_NoStatusChange(t *testing.T) {
	existingJob := domain.NewJob("task", nil, 1, nil)
	repo := &mockJobRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Job, error) {
			if id == existingJob.ID() {
				return existingJob, nil
			}
			return nil, domain.ErrJobNotFound
		},
	}
	eb := &mockEventBus{}
	h := NewUpdateJobHandler(repo, eb, &mockLogger{})

	// No status provided
	cmd := UpdateJobCommand{
		ID: existingJob.ID(),
	}

	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.updatedJob.Status() != domain.JobStatusPending {
		t.Fatalf("expected status to remain PENDING, got %s", repo.updatedJob.Status())
	}
}

// ---------------------------------------------------------------------------
// DeleteJobHandler tests
// ---------------------------------------------------------------------------

func TestDeleteJobHandler_Success(t *testing.T) {
	jobID := uuid.New()
	repo := &mockJobRepo{}
	h := NewDeleteJobHandler(repo, &mockLogger{})

	cmd := DeleteJobCommand{ID: jobID}
	err := h.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.deletedID != jobID {
		t.Fatalf("expected deleted ID %s, got %s", jobID, repo.deletedID)
	}
}

func TestDeleteJobHandler_RepoError(t *testing.T) {
	repoErr := errors.New("delete failed")
	repo := &mockJobRepo{
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return repoErr
		},
	}
	h := NewDeleteJobHandler(repo, &mockLogger{})

	cmd := DeleteJobCommand{ID: uuid.New()}
	err := h.Handle(context.Background(), cmd)
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
