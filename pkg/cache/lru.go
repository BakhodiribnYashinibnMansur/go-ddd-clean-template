package cache

import (
	"container/list"
	"errors"
	"sync"
)

// ErrInvalidCapacity is returned when cache capacity is invalid.
var ErrInvalidCapacity = errors.New("capacity must be greater than 0")

// LRUCache implements a Least Recently Used cache
type LRUCache struct {
	capacity int
	items    map[string]*list.Element
	list     *list.List
	mu       sync.RWMutex
}

type lruEntry struct {
	key   string
	value any
}

// NewLRUCache creates a new LRU cache with the given capacity
func NewLRUCache(capacity int) *LRUCache {
	if capacity <= 0 {
		capacity = 100
	}
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

// Set adds or updates a value in the cache
func (c *LRUCache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.list.MoveToFront(elem)
		if entry, ok := elem.Value.(*lruEntry); ok {
			entry.value = value
		}
		return
	}

	if c.list.Len() >= c.capacity {
		c.evict()
	}

	entry := &lruEntry{key: key, value: value}
	elem := c.list.PushFront(entry)
	c.items[key] = elem
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.list.MoveToFront(elem)
		if entry, ok := elem.Value.(*lruEntry); ok {
			return entry.value, true
		}
	}
	return nil, false
}

// Remove removes a value from the cache
func (c *LRUCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.removeElement(elem)
	}
}

func (c *LRUCache) evict() {
	elem := c.list.Back()
	if elem != nil {
		c.removeElement(elem)
	}
}

func (c *LRUCache) removeElement(elem *list.Element) {
	c.list.Remove(elem)
	if entry, ok := elem.Value.(*lruEntry); ok {
		delete(c.items, entry.key)
	}
}

// Len returns the number of items in the cache
func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.list.Len()
}

// Purge clears the cache
func (c *LRUCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*list.Element)
	c.list.Init()
}
