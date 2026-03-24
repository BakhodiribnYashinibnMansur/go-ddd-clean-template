package metric_test

import (
	"context"

	"gct/config"
	"gct/internal/domain"
	"gct/internal/usecase"
	ucaudit "gct/internal/usecase/audit"
	ucauditlog "gct/internal/usecase/audit/auditlog"
	ucendpointhistory "gct/internal/usecase/audit/endpointhistory"
	ucmetric "gct/internal/usecase/audit/metric"
	ucsystemerror "gct/internal/usecase/audit/systemerror"

	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// MockMetricUseCase implements ucmetric.UseCaseI
// ---------------------------------------------------------------------------

type MockMetricUseCase struct {
	mock.Mock
}

func (m *MockMetricUseCase) Create(ctx context.Context, in *domain.FunctionMetric) error {
	return m.Called(ctx, in).Error(0)
}

func (m *MockMetricUseCase) Gets(ctx context.Context, in *domain.FunctionMetricsFilter) ([]*domain.FunctionMetric, int, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*domain.FunctionMetric), args.Get(1).(int), args.Error(2)
}

func (m *MockMetricUseCase) MeasureSafe(ctx context.Context, name string) func() {
	return func() {}
}

// ---------------------------------------------------------------------------
// Stub mocks for other audit sub-usecases
// ---------------------------------------------------------------------------

type MockAuditLogUseCase struct{ mock.Mock }

func (m *MockAuditLogUseCase) Create(ctx context.Context, in *domain.AuditLog) error {
	return m.Called(ctx, in).Error(0)
}
func (m *MockAuditLogUseCase) Gets(ctx context.Context, in *domain.AuditLogsFilter) ([]*domain.AuditLog, int, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*domain.AuditLog), args.Get(1).(int), args.Error(2)
}
func (m *MockAuditLogUseCase) GetLogins(ctx context.Context, in *domain.AuditLogsFilter) ([]domain.LoginEntry, int, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]domain.LoginEntry), args.Get(1).(int), args.Error(2)
}
func (m *MockAuditLogUseCase) GetSessions(ctx context.Context, in *domain.AuditLogsFilter) ([]domain.SessionEntry, int, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]domain.SessionEntry), args.Get(1).(int), args.Error(2)
}
func (m *MockAuditLogUseCase) GetActions(ctx context.Context, in *domain.AuditLogsFilter) ([]domain.ActionEntry, int, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]domain.ActionEntry), args.Get(1).(int), args.Error(2)
}

type MockEndpointHistoryUseCase struct{ mock.Mock }

func (m *MockEndpointHistoryUseCase) Create(ctx context.Context, in *domain.EndpointHistory) error {
	return m.Called(ctx, in).Error(0)
}
func (m *MockEndpointHistoryUseCase) Gets(ctx context.Context, in *domain.EndpointHistoriesFilter) ([]*domain.EndpointHistory, int, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*domain.EndpointHistory), args.Get(1).(int), args.Error(2)
}

type MockSystemErrorUseCase struct{ mock.Mock }

func (m *MockSystemErrorUseCase) Create(ctx context.Context, in *domain.SystemError) error {
	return m.Called(ctx, in).Error(0)
}
func (m *MockSystemErrorUseCase) Gets(ctx context.Context, in *domain.SystemErrorsFilter) ([]*domain.SystemError, int, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int), args.Error(2)
	}
	return args.Get(0).([]*domain.SystemError), args.Get(1).(int), args.Error(2)
}
func (m *MockSystemErrorUseCase) Resolve(ctx context.Context, id string, resolvedBy *string) error {
	return m.Called(ctx, id, resolvedBy).Error(0)
}

// ---------------------------------------------------------------------------
// MockAuditUseCase implements ucaudit.UseCaseI
// ---------------------------------------------------------------------------

type MockAuditUseCase struct {
	logUC         ucauditlog.UseCaseI
	historyUC     ucendpointhistory.UseCaseI
	metricUC      ucmetric.UseCaseI
	systemErrorUC ucsystemerror.UseCaseI
}

func (m *MockAuditUseCase) Log() ucauditlog.UseCaseI            { return m.logUC }
func (m *MockAuditUseCase) History() ucendpointhistory.UseCaseI  { return m.historyUC }
func (m *MockAuditUseCase) Metric() ucmetric.UseCaseI            { return m.metricUC }
func (m *MockAuditUseCase) SystemError() ucsystemerror.UseCaseI  { return m.systemErrorUC }

// ---------------------------------------------------------------------------
// Helper: build a *usecase.UseCase with mocked Audit.Metric sub-service
// ---------------------------------------------------------------------------

func buildUseCase(metricMock *MockMetricUseCase) *usecase.UseCase {
	return &usecase.UseCase{
		Audit: &MockAuditUseCase{
			logUC:         new(MockAuditLogUseCase),
			historyUC:     new(MockEndpointHistoryUseCase),
			metricUC:      metricMock,
			systemErrorUC: new(MockSystemErrorUseCase),
		},
	}
}

func buildConfig() *config.Config {
	return &config.Config{}
}

// Ensure interfaces are satisfied at compile time.
var _ ucmetric.UseCaseI = (*MockMetricUseCase)(nil)
var _ ucauditlog.UseCaseI = (*MockAuditLogUseCase)(nil)
var _ ucendpointhistory.UseCaseI = (*MockEndpointHistoryUseCase)(nil)
var _ ucsystemerror.UseCaseI = (*MockSystemErrorUseCase)(nil)
var _ ucaudit.UseCaseI = (*MockAuditUseCase)(nil)
