package outbox

import (
	"encoding/json"
	"fmt"

	shareddomain "gct/internal/platform/domain"

	"github.com/google/uuid"
)

// ToEntries serializes a batch of domain events into outbox entries ready to
// be appended within a transaction. Callers should supply the same
// Published Language contracts (gct/internal/contract/events) that will
// appear on the bus — the outbox row's payload is a JSON encoding of the
// event value.
func ToEntries(events []shareddomain.DomainEvent) ([]Entry, error) {
	out := make([]Entry, 0, len(events))
	for _, ev := range events {
		raw, err := json.Marshal(ev)
		if err != nil {
			return nil, fmt.Errorf("outbox marshal %s: %w", ev.EventName(), err)
		}
		out = append(out, Entry{
			EventID:     uuid.New(), // envelope carries its own EventID in payload; this row id is just for dedupe
			EventName:   ev.EventName(),
			AggregateID: ev.AggregateID(),
			Payload:     raw,
			OccurredAt:  ev.OccurredAt(),
		})
	}
	return out, nil
}
