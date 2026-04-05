// Package published maps the user BC's internal domain events to its
// Published Language contracts in gct/internal/contract/events. This
// package is the ONLY place in the user BC that knows about the wire
// format; the domain layer stays free of integration concerns and
// consumers never import the user BC.
package published

import (
	"gct/internal/context/iam/user/domain"
	"gct/internal/contract/events"
	shareddomain "gct/internal/platform/domain"
)

// Map translates a batch of internal user domain events into the stable
// versioned contracts. Unknown events are dropped. Call after the
// aggregate's transaction commits, then hand the result to the event bus
// or the transactional outbox.
func Map(internal []shareddomain.DomainEvent) []shareddomain.DomainEvent {
	out := make([]shareddomain.DomainEvent, 0, len(internal))
	for _, e := range internal {
		switch v := e.(type) {
		case domain.UserCreated:
			out = append(out, events.NewUserCreatedV1(v.AggregateID(), v.Phone))
		case domain.UserSignedIn:
			out = append(out, events.NewUserSignedInV1(v.AggregateID(), v.SessionID, v.IPAddress))
		case domain.UserDeactivated:
			out = append(out, events.NewUserDeactivatedV1(v.AggregateID()))
		case domain.PasswordChanged:
			out = append(out, events.NewUserPasswordChangedV1(v.AggregateID()))
		case domain.UserApproved:
			out = append(out, events.NewUserApprovedV1(v.AggregateID()))
		case domain.RoleChanged:
			out = append(out, events.NewUserRoleChangedV1(v.AggregateID(), v.OldRoleID, v.NewRoleID))
		}
	}
	return out
}
