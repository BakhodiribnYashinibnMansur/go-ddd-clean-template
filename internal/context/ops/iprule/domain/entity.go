package domain

import (
	"time"

	shared "gct/internal/platform/domain"

	"github.com/google/uuid"
)

// IPRule is the aggregate root for IP-based access control rules.
// Each rule maps a single IP address to an ALLOW or DENY action, with an optional expiration.
// Once expired, the rule should be treated as inactive — enforcement logic must check expiresAt.
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

// Update applies a partial update to the IP rule using pointer-based optionality.
// Callers should re-validate action values (ALLOW/DENY) before calling this method,
// as the aggregate does not enforce the enum constraint internally.
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
