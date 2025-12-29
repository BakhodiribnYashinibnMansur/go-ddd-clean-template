package cache

import (
	"sync"
)

// LIFOCache implements a Last-In-First-Out cache (Stack based)
type LIFOCache struct {
	capacity int
	items    map[string]any
	stack    []string
	mu       sync.RWMutex
}

// NewLIFOCache creates a new LIFO cache
func NewLIFOCache(capacity int) *LIFOCache {
	return &LIFOCache{
		capacity: capacity,
		items:    make(map[string]any),
		stack:    make([]string, 0, capacity),
	}
}

// Set adds a value to the cache
func (c *LIFOCache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If exists, update (LIFO policy usually doesn't change position on update,
	// or it strictly treats it as a stack. We'll just update value here)
	if _, exists := c.items[key]; exists {
		c.items[key] = value
		return
	}

	// Evict if full
	if len(c.stack) >= c.capacity {
		c.evict()
	}

	c.items[key] = value
	c.stack = append(c.stack, key)
}

// Get retrieves a value
func (c *LIFOCache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.items[key]
	return val, ok
}

// Remove removes a value
func (c *LIFOCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.items[key]; !ok {
		return
	}

	delete(c.items, key)
	// Rebuild stack (expensive O(n), but simple for LIFO)
	// For production LIFO, a doubly linked list is better, but slice is fine for demo
	newStack := make([]string, 0, len(c.stack)-1)
	for _, k := range c.stack {
		if k != key {
			newStack = append(newStack, k)
		}
	}
	c.stack = newStack
}

// evict removes the most recently added value (Top of stack)
func (c *LIFOCache) evict() {
	if len(c.stack) == 0 {
		return
	}
	lastIdx := len(c.stack) - 1
	key := c.stack[lastIdx]
	c.stack = c.stack[:lastIdx]
	delete(c.items, key)
}

// Len returns number of items
func (c *LIFOCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Purge clears the cache
func (c *LIFOCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]any)
	c.stack = make([]string, 0, c.capacity)
}
