package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// Translation is the aggregate root for i18n key-value pairs.
// Each instance represents a single key in a single language; the (key, language) pair forms the natural uniqueness constraint.
// The group field enables logical grouping (e.g., "auth", "dashboard") for bulk export or frontend module loading.
type Translation struct {
	shared.AggregateRoot
	key      string
	language string
	value    string
	group    string
}

// NewTranslation creates a new Translation aggregate.
func NewTranslation(key, language, value, group string) *Translation {
	return &Translation{
		AggregateRoot: shared.NewAggregateRoot(),
		key:           key,
		language:      language,
		value:         value,
		group:         group,
	}
}

// ReconstructTranslation rebuilds a Translation from persisted data. No events are raised.
func ReconstructTranslation(
	id uuid.UUID,
	createdAt, updatedAt time.Time,
	key, language, value, group string,
) *Translation {
	return &Translation{
		AggregateRoot: shared.NewAggregateRootWithID(id, createdAt, updatedAt, nil),
		key:           key,
		language:      language,
		value:         value,
		group:         group,
	}
}

// Update applies partial modifications using pointer semantics — nil fields are left unchanged.
// Changing the key or language effectively re-identifies the translation; callers should
// ensure no duplicate (key, language) pair exists before calling this.
func (t *Translation) Update(key, language, value, group *string) {
	if key != nil {
		t.key = *key
	}
	if language != nil {
		t.language = *language
	}
	if value != nil {
		t.value = *value
	}
	if group != nil {
		t.group = *group
	}
	t.Touch()
	t.AddEvent(NewTranslationUpdated(t.ID()))
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (t *Translation) Key() string      { return t.key }
func (t *Translation) Language() string  { return t.language }
func (t *Translation) Value() string     { return t.value }
func (t *Translation) Group() string     { return t.group }
