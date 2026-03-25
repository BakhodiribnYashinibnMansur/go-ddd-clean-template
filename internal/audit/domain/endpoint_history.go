package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// EndpointHistory tracks HTTP request history for auditing purposes.
type EndpointHistory struct {
	shared.BaseEntity
	userID     *uuid.UUID
	endpoint   string
	method     string
	statusCode int
	latency    int
	ipAddress  *string
	userAgent  *string
	createdAt  time.Time
}

// NewEndpointHistory creates a new EndpointHistory entity.
func NewEndpointHistory(
	userID *uuid.UUID,
	endpoint string,
	method string,
	statusCode int,
	latency int,
	ipAddress *string,
	userAgent *string,
) *EndpointHistory {
	now := time.Now()
	return &EndpointHistory{
		BaseEntity: shared.NewBaseEntity(),
		userID:     userID,
		endpoint:   endpoint,
		method:     method,
		statusCode: statusCode,
		latency:    latency,
		ipAddress:  ipAddress,
		userAgent:  userAgent,
		createdAt:  now,
	}
}

// ReconstructEndpointHistory rebuilds an EndpointHistory from persisted data.
func ReconstructEndpointHistory(
	id uuid.UUID,
	createdAt time.Time,
	userID *uuid.UUID,
	endpoint string,
	method string,
	statusCode int,
	latency int,
	ipAddress *string,
	userAgent *string,
) *EndpointHistory {
	return &EndpointHistory{
		BaseEntity: shared.NewBaseEntityWithID(id, createdAt, createdAt, nil),
		userID:     userID,
		endpoint:   endpoint,
		method:     method,
		statusCode: statusCode,
		latency:    latency,
		ipAddress:  ipAddress,
		userAgent:  userAgent,
		createdAt:  createdAt,
	}
}

// Getters

func (e *EndpointHistory) UserID() *uuid.UUID    { return e.userID }
func (e *EndpointHistory) Endpoint() string       { return e.endpoint }
func (e *EndpointHistory) Method() string         { return e.method }
func (e *EndpointHistory) StatusCode() int        { return e.statusCode }
func (e *EndpointHistory) Latency() int           { return e.latency }
func (e *EndpointHistory) IPAddress() *string      { return e.ipAddress }
func (e *EndpointHistory) UserAgent() *string      { return e.userAgent }
func (e *EndpointHistory) GetCreatedAt() time.Time { return e.createdAt }
