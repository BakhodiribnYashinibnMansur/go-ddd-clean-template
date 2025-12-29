package cache

import (
	"container/list"
	"sync"
)

// SLRUCache implements Segmented LRU
type SLRUCache struct {
	capacity     int
	probationCap int
	protectedCap int
	items        map[string]*list.Element
	probationary *list.List
	protected    *list.List
	mu           sync.RWMutex
}

type slruEntry struct {
	key         string
	value       any
	isProtected bool
}

// NewSLRUCache creates a new SLRU cache
// capacity is total capacity. ratio is percent of protected segment (e.g., 0.8)
func NewSLRUCache(capacity int, protectedRatio float64) *SLRUCache {
	protectedCap := int(float64(capacity) * protectedRatio)
	if protectedCap < 1 {
		protectedCap = 1
	}
	probationCap := capacity - protectedCap
	if probationCap < 1 {
		probationCap = 1 // Ensure at least some probation
	}

	return &SLRUCache{
		capacity:     capacity,
		probationCap: probationCap,
		protectedCap: protectedCap,
		items:        make(map[string]*list.Element),
		probationary: list.New(),
		protected:    list.New(),
	}
}

// Set adds a value
func (c *SLRUCache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		entry := elem.Value.(*slruEntry) //nolint:forcetypeassert // safe: we control the type
		entry.value = value
		// Hit logic
		if entry.isProtected {
			c.protected.MoveToFront(elem)
		} else {
			// Promote from Probationary to Protected
			c.probationary.Remove(elem)
			entry.isProtected = true
			newElem := c.protected.PushFront(entry)
			c.items[key] = newElem
			c.ensureCapacity()
		}
		return
	}

	// New item -> Probationary MRU
	if c.probationary.Len() >= c.probationCap {
		// If probation is full, we must evict from probation to make room
		// But wait, ensureCapacity handles overall logic?
		// Simpler: just evict logic right here specific for insert
		c.evictProbationary()
	}

	entry := &slruEntry{key: key, value: value, isProtected: false}
	elem := c.probationary.PushFront(entry)
	c.items[key] = elem
}

// Get retrieves a value
func (c *SLRUCache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return nil, false
	}

	entry := elem.Value.(*slruEntry) //nolint:forcetypeassert // safe: we control the type
	if entry.isProtected {
		c.protected.MoveToFront(elem)
	} else {
		// Promote
		c.probationary.Remove(elem)
		entry.isProtected = true
		newElem := c.protected.PushFront(entry)
		c.items[key] = newElem
		c.ensureCapacity()
	}

	return entry.value, true
}

func (c *SLRUCache) ensureCapacity() {
	// If protected is full, demote LRU of protected to MRU of probationary
	if c.protected.Len() > c.protectedCap {
		back := c.protected.Back()
		if back != nil {
			c.protected.Remove(back)
			entry := back.Value.(*slruEntry) //nolint:forcetypeassert // safe: we control the type
			entry.isProtected = false

			// Move to probationary
			newElem := c.probationary.PushFront(entry)
			c.items[entry.key] = newElem

			// If probationary is now full, evict from it
			if c.probationary.Len() > c.probationCap {
				c.evictProbationary()
			}
		}
	}
}

func (c *SLRUCache) evictProbationary() {
	back := c.probationary.Back()
	if back != nil {
		c.probationary.Remove(back)
		entry := back.Value.(*slruEntry) //nolint:forcetypeassert // safe: we control the type
		delete(c.items, entry.key)
	}
}

// Remove removes a value
func (c *SLRUCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[key]
	if !ok {
		return
	}
	entry := elem.Value.(*slruEntry) //nolint:forcetypeassert // safe: we control the type
	if entry.isProtected {
		c.protected.Remove(elem)
	} else {
		c.probationary.Remove(elem)
	}
	delete(c.items, key)
}

// Len returns size
func (c *SLRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Purge clears cache
func (c *SLRUCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*list.Element)
	c.probationary.Init()
	c.protected.Init()
}
