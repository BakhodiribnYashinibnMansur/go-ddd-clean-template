package cache

import (
	"crypto/rand"
	"math/big"
	"sync"
)

// RandomCache implements a Random Eviction cache
type RandomCache struct {
	capacity int
	items    map[string]any
	keys     []string
	mu       sync.RWMutex
}

// NewRandomCache creates a new Random cache
func NewRandomCache(capacity int) *RandomCache {
	return &RandomCache{
		capacity: capacity,
		items:    make(map[string]any),
		keys:     make([]string, 0, capacity),
	}
}

// Set adds a value
func (c *RandomCache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[key]; exists {
		c.items[key] = value
		return
	}

	if len(c.keys) >= c.capacity {
		c.evict()
	}

	c.items[key] = value
	c.keys = append(c.keys, key)
}

// Get retrieves a value
func (c *RandomCache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.items[key]
	return val, ok
}

// Remove removes a value
func (c *RandomCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.items[key]; !ok {
		return
	}
	delete(c.items, key)

	// O(n) removal from slice
	for i, k := range c.keys {
		if k == key {
			// Swap with last and truncate
			c.keys[i] = c.keys[len(c.keys)-1]
			c.keys = c.keys[:len(c.keys)-1]
			break
		}
	}
}

// evict removes a random item
func (c *RandomCache) evict() {
	if len(c.keys) == 0 {
		return
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(c.keys))))
	if err != nil {
		// Fallback to 0 if rand fails, which shouldn't happen
		n = big.NewInt(0)
	}
	idx := int(n.Int64())
	key := c.keys[idx] // This is the Victim

	// Swap delete
	c.keys[idx] = c.keys[len(c.keys)-1]
	c.keys = c.keys[:len(c.keys)-1]

	delete(c.items, key)
}

// Len returns size
func (c *RandomCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Purge clears cache
func (c *RandomCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]any)
	c.keys = make([]string, 0, c.capacity)
}
