package mock

import (
	"time"

	"gct/internal/domain"
	"github.com/brianvoe/gofakeit/v7"
)

// AuditLog generates a fake domain.AuditLog
func AuditLog() *domain.AuditLog {
	userID := UUID()
	sessionID := UUID()
	resourceID := UUID()
	resourceType := gofakeit.Noun()
	platform := gofakeit.AppName()
	ip := gofakeit.IPv4Address()
	ua := gofakeit.UserAgent()
	permission := gofakeit.Verb()
	policyID := UUID()
	decision := gofakeit.Verb()
	errMsg := gofakeit.Sentence(5)

	return &domain.AuditLog{
		ID:           UUID(),
		UserID:       &userID,
		SessionID:    &sessionID,
		Action:       domain.AuditActionLogin,
		ResourceType: &resourceType,
		ResourceID:   &resourceID,
		Platform:     &platform,
		IPAddress:    &ip,
		UserAgent:    &ua,
		Permission:   &permission,
		PolicyID:     &policyID,
		Decision:     &decision,
		Success:      gofakeit.Bool(),
		ErrorMessage: &errMsg,
		Metadata:     map[string]any{"key": gofakeit.Word()},
		CreatedAt:    time.Now(),
	}
}

// AuditLogs generates multiple fake domain.AuditLog
func AuditLogs(count int) []*domain.AuditLog {
	logs := make([]*domain.AuditLog, count)
	for i := range count {
		logs[i] = AuditLog()
	}
	return logs
}

// EndpointHistory generates a fake domain.EndpointHistory
func EndpointHistory() *domain.EndpointHistory {
	userID := UUID()
	sessionID := UUID()
	platform := gofakeit.AppName()
	ip := gofakeit.IPv4Address()
	ua := gofakeit.UserAgent()
	permission := gofakeit.Verb()
	decision := gofakeit.Verb()
	requestID := UUID()
	errMsg := gofakeit.Sentence(5)
	respSize := gofakeit.IntRange(100, 10000)

	return &domain.EndpointHistory{
		ID:           UUID(),
		UserID:       &userID,
		SessionID:    &sessionID,
		Method:       gofakeit.HTTPMethod(),
		Path:         gofakeit.URL(),
		StatusCode:   gofakeit.HTTPStatusCode(),
		DurationMs:   gofakeit.IntRange(1, 1000),
		Platform:     &platform,
		IPAddress:    &ip,
		UserAgent:    &ua,
		Permission:   &permission,
		Decision:     &decision,
		RequestID:    &requestID,
		RateLimited:  gofakeit.Bool(),
		ResponseSize: &respSize,
		ErrorMessage: &errMsg,
		CreatedAt:    time.Now(),
	}
}

// EndpointHistories generates multiple fake domain.EndpointHistory
func EndpointHistories(count int) []*domain.EndpointHistory {
	histories := make([]*domain.EndpointHistory, count)
	for i := range count {
		histories[i] = EndpointHistory()
	}
	return histories
}

// FunctionMetric generates a fake domain.FunctionMetric
func FunctionMetric() *domain.FunctionMetric {
	errMsg := gofakeit.Sentence(5)
	return &domain.FunctionMetric{
		ID:         UUID(),
		Name:       gofakeit.AppName(),
		LatencyMs:  gofakeit.IntRange(1, 1000),
		IsPanic:    gofakeit.Bool(),
		PanicError: &errMsg,
		CreatedAt:  time.Now(),
	}
}

// FunctionMetrics generates multiple fake domain.FunctionMetric
func FunctionMetrics(count int) []*domain.FunctionMetric {
	metrics := make([]*domain.FunctionMetric, count)
	for i := range count {
		metrics[i] = FunctionMetric()
	}
	return metrics
}
