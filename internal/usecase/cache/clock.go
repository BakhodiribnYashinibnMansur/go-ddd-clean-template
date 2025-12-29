package cache

import (
	"sync"
)

// ClockCache implements the Clock (Second Chance) algorithm
type ClockCache struct {
	capacity int
	items    map[string]*clockEntry
	// circular buffer of keys, or pointers to entries
	// We'll use a slice of keys for simplicity in 'circle'
	circle []*clockEntry
	hand   int
	mu     sync.RWMutex
}

type clockEntry struct {
	key    string
	value  any
	refBit bool
}

// NewClockCache creates a new Clock cache
func NewClockCache(capacity int) *ClockCache {
	return &ClockCache{
		capacity: capacity,
		items:    make(map[string]*clockEntry),
		circle:   make([]*clockEntry, 0, capacity),
		hand:     0,
	}
}

// Set adds a value
func (c *ClockCache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.items[key]; ok {
		entry.value = value
		entry.refBit = true
		return
	}

	if len(c.items) >= c.capacity {
		c.evict()
	}

	entry := &clockEntry{
		key:    key,
		value:  value,
		refBit: true, // Initially gave it a chance? Usually yes.
	}
	c.items[key] = entry
	c.circle = append(c.circle, entry) // Not strictly circular in slice, but evict will maintain size logic
	// Actually, strictly correct clock requires fixed slots effectively.
	// If append grows, 'hand' logic gets complex.
	// Easier: if Full, 'evict' replaces an item in place in 'circle'.
	// If Not Full, we just append.
}

// Get retrieves a value
func (c *ClockCache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.items[key]; ok {
		entry.refBit = true
		return entry.value, true
	}
	return nil, false
}

func (c *ClockCache) evict() {
	// Clock algorithm
	// Iterate starting from hand
	for {
		if c.hand >= len(c.circle) {
			c.hand = 0
		}

		entry := c.circle[c.hand]
		if entry.refBit {
			entry.refBit = false
			c.hand++
		} else {
			// Evict this one
			delete(c.items, entry.key)

			// We remove it from the circle?
			// Standard Clock overwrites the slot.
			// But Set() creates a new Entry.
			// So effectively we remove this key from map,
			// and Remove this entry from circle. Or replace logic.

			// For simpler impelemntation where Set() is separate:
			// Remove from circle
			c.removeAtHand()
			return
		}
	}
}

func (c *ClockCache) removeAtHand() {
	// Remove element at c.hand
	copy(c.circle[c.hand:], c.circle[c.hand+1:])
	c.circle[len(c.circle)-1] = nil
	c.circle = c.circle[:len(c.circle)-1]

	// Hand stays at same index (which now points to next element),
	// unless we were at end
	if c.hand >= len(c.circle) {
		c.hand = 0
	}
}

// Remove removes a value
func (c *ClockCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.items[key]; !ok {
		return
	}
	delete(c.items, key)

	// Find and remove from circle O(N)
	for i, entry := range c.circle {
		if entry.key == key {
			copy(c.circle[i:], c.circle[i+1:])
			c.circle[len(c.circle)-1] = nil
			c.circle = c.circle[:len(c.circle)-1]

			// Adjust hand if needed
			if i < c.hand {
				c.hand--
			}
			break
		}
	}
}

// Len returns size
func (c *ClockCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Purge clears cache
func (c *ClockCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*clockEntry)
	c.circle = make([]*clockEntry, 0, c.capacity)
	c.hand = 0
}
