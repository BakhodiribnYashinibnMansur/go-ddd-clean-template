package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// IPRule is the aggregate root for IP rules.
type IPRule struct {
	shared.AggregateRoot
	ipAddress string
	action    string // ALLOW or DENY
	reason    string
	expiresAt *time.Time
}

// NewIPRule creates a new IPRule aggregate and raises an IPRuleCreated event.
func NewIPRule(ipAddress, action, reason string, expiresAt *time.Time) *IPRule {
	r := &IPRule{
		AggregateRoot: shared.NewAggregateRoot(),
		ipAddress:     ipAddress,
		action:        action,
		reason:        reason,
		expiresAt:     expiresAt,
	}
	r.AddEvent(NewIPRuleCreated(r.ID(), ipAddress, action))
	return r
}

// ReconstructIPRule rebuilds an IPRule from persisted data. No events are raised.
func ReconstructIPRule(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	ipAddress, action, reason string,
	expiresAt *time.Time,
) *IPRule {
	return &IPRule{
		AggregateRoot: shared.NewAggregateRootWithID(id, createdAt, updatedAt, nil),
		ipAddress:     ipAddress,
		action:        action,
		reason:        reason,
		expiresAt:     expiresAt,
	}
}

// Update modifies the IP rule fields.
func (r *IPRule) Update(ipAddress, action, reason *string, expiresAt *time.Time) {
	if ipAddress != nil {
		r.ipAddress = *ipAddress
	}
	if action != nil {
		r.action = *action
	}
	if reason != nil {
		r.reason = *reason
	}
	if expiresAt != nil {
		r.expiresAt = expiresAt
	}
	r.Touch()
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (r *IPRule) IPAddress() string     { return r.ipAddress }
func (r *IPRule) Action() string        { return r.action }
func (r *IPRule) Reason() string        { return r.reason }
func (r *IPRule) ExpiresAt() *time.Time { return r.expiresAt }
