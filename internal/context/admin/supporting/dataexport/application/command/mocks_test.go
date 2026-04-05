package command_test

import (
	"context"

	"gct/internal/context/admin/supporting/dataexport/domain"
	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"
)

// ---------------------------------------------------------------------------
// Mock infrastructure
// ---------------------------------------------------------------------------

type mockWriteRepo struct {
	savedEntity   *domain.DataExport
	updatedEntity *domain.DataExport
	deletedID     domain.DataExportID
	findByIDFn    func(ctx context.Context, id domain.DataExportID) (*domain.DataExport, error)
	saveFn        func(ctx context.Context, entity *domain.DataExport) error
	updateFn      func(ctx context.Context, entity *domain.DataExport) error
	deleteFn      func(ctx context.Context, id domain.DataExportID) error
}

func (m *mockWriteRepo) Save(ctx context.Context, entity *domain.DataExport) error {
	m.savedEntity = entity
	if m.saveFn != nil {
		return m.saveFn(ctx, entity)
	}
	return nil
}

func (m *mockWriteRepo) FindByID(ctx context.Context, id domain.DataExportID) (*domain.DataExport, error) {
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

func (m *mockWriteRepo) Delete(ctx context.Context, id domain.DataExportID) error {
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

func (m *mockEventBus) Subscribe(_ string, _ application.EventHandler) error {
	return nil
}

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
