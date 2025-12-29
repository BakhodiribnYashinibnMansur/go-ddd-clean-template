package cache

import (
	"container/list"
	"sync"
)

// MRUCache implements a Most-Recently-Used cache
type MRUCache struct {
	capacity int
	data     map[string]*list.Element
	list     *list.List
	mu       sync.RWMutex
}

type mruEntry struct {
	key   string
	value any
}

// NewMRUCache creates a new MRU cache
func NewMRUCache(capacity int) *MRUCache {
	return &MRUCache{
		capacity: capacity,
		data:     make(map[string]*list.Element),
		list:     list.New(),
	}
}

// Set adds or updates a value
func (c *MRUCache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.data[key]; ok {
		// Update value and move to Front (Most Recently Used)
		elem.Value.(*mruEntry).value = value //nolint:forcetypeassert // safe: we control the type
		c.list.MoveToFront(elem)
		return
	}

	if c.list.Len() >= c.capacity {
		c.evict()
	}

	entry := &mruEntry{key: key, value: value}
	elem := c.list.PushFront(entry)
	c.data[key] = elem
}

// Get retrieves a value
func (c *MRUCache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.data[key]; ok {
		c.list.MoveToFront(elem)
		return elem.Value.(*mruEntry).value, true //nolint:forcetypeassert // safe: we control the type
	}
	return nil, false
}

// Remove removes a value
func (c *MRUCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.data[key]; ok {
		c.removeElement(elem)
	}
}

// evict removes the Most Recently Used item (Front of list)
func (c *MRUCache) evict() {
	elem := c.list.Front()
	if elem != nil {
		c.removeElement(elem)
	}
}

func (c *MRUCache) removeElement(elem *list.Element) {
	c.list.Remove(elem)
	entry := elem.Value.(*mruEntry) //nolint:forcetypeassert // safe: we control the type
	delete(c.data, entry.key)
}

// Len returns size
func (c *MRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.list.Len()
}

// Purge clears cache
func (c *MRUCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]*list.Element)
	c.list.Init()
}
