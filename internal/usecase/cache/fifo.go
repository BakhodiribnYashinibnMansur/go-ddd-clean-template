package cache

import (
	"container/list"
	"sync"
)

// FIFOCache implements a First-In-First-Out cache
type FIFOCache struct {
	capacity int
	data     map[string]*list.Element
	queue    *list.List
	mu       sync.RWMutex
}

type fifoEntry struct {
	key   string
	value any
}

// NewFIFOCache creates a new FIFO cache with the given capacity
func NewFIFOCache(capacity int) *FIFOCache {
	return &FIFOCache{
		capacity: capacity,
		data:     make(map[string]*list.Element),
		queue:    list.New(),
	}
}

// Set adds a value to the cache
func (c *FIFOCache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.data[key]; ok {
		// Update value, but don't move it (FIFO doesn't care about access)
		elem.Value.(*fifoEntry).value = value //nolint:forcetypeassert // safe: we control the type
		return
	}

	entry := &fifoEntry{key: key, value: value}
	elem := c.queue.PushBack(entry)
	c.data[key] = elem

	if c.queue.Len() > c.capacity {
		c.evict()
	}
}

// Get retrieves a value from the cache
func (c *FIFOCache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if elem, ok := c.data[key]; ok {
		return elem.Value.(*fifoEntry).value, true //nolint:forcetypeassert // safe: we control the type
	}
	return nil, false
}

// Remove removes a value from the cache
func (c *FIFOCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.data[key]; ok {
		c.removeElement(elem)
	}
}

// evict removes the oldest item
func (c *FIFOCache) evict() {
	elem := c.queue.Front()
	if elem != nil {
		c.removeElement(elem)
	}
}

func (c *FIFOCache) removeElement(elem *list.Element) {
	c.queue.Remove(elem)
	entry := elem.Value.(*fifoEntry) //nolint:forcetypeassert // safe: we control the type
	delete(c.data, entry.key)
}

// Len returns the number of items in the cache
func (c *FIFOCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.queue.Len()
}

// Purge clear the cache
func (c *FIFOCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]*list.Element)
	c.queue.Init()
}
