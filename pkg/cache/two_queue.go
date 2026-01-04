package cache

import (
	"container/list"
	"sync"
)

type qType int

const (
	inQ1 qType = iota
	inQ2
	inQout
)

// TwoQueueCache implements the 2Q cache algorithm as per the provided Technical Specification (TZ).
// It uses three queues: A1in (Q1), Am (Q2), and A1out (Qout) to manage cache entries.
type TwoQueueCache struct {
	maxSize       int
	a1InCapacity  int
	amCapacity    int
	a1OutCapacity int

	// Structures
	q1   *list.List // A1in: new items
	q2   *list.List // Am: hot items
	qout *list.List // A1out: evicted keys history (ghost)

	items    map[string]*list.Element // Elements in q1 and q2
	outItems map[string]*list.Element // Elements in qout (keys only)

	// Metrics
	hits           int
	misses         int
	promotionCount int

	mu sync.RWMutex
}

type twoQueueEntry struct {
	key   string
	value any
	q     qType
}

// NewTwoQueueCache creates a new 2Q cache based on the provided technical specification.
// maxSize: Total number of items allowed in memory (A1in + Am).
func NewTwoQueueCache(maxSize int) *TwoQueueCache {
	if maxSize <= 0 {
		maxSize = 100 // Default or handle error
	}
	a1InCap := maxSize / 4 // 25%
	if a1InCap < 1 {
		a1InCap = 1
	}
	amCap := maxSize - a1InCap // 75%
	if amCap < 1 {
		amCap = 1
	}

	return &TwoQueueCache{
		maxSize:       maxSize,
		a1InCapacity:  a1InCap,
		amCapacity:    amCap,
		a1OutCapacity: maxSize, // 100% per TZ

		q1:   list.New(),
		q2:   list.New(),
		qout: list.New(),

		items:    make(map[string]*list.Element),
		outItems: make(map[string]*list.Element),
	}
}

// Get retrieves an item from the cache.
// Following the 2Q algorithm:
// - If in Am (Q2): promote to MRU in Q2.
// - If in A1in (Q1): promote from Q1 to Q2.
// - If in A1out (Qout): count as miss.
func (c *TwoQueueCache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check Am (Q2) or A1in (Q1)
	if elem, ok := c.items[key]; ok {
		entry := elem.Value.(*twoQueueEntry) // safe: we control the type

		// If it's already in Q2, move to MRU (front) of Q2
		if entry.q == inQ2 {
			c.q2.MoveToFront(elem)
			c.hits++
			return entry.value, true
		}

		// If it's in Q1, it's being accessed again, so promote to Q2 (Am)
		if entry.q == inQ1 {
			c.q1.Remove(elem)
			entry.q = inQ2
			newElem := c.q2.PushFront(entry)
			c.items[key] = newElem
			c.hits++
			c.promotionCount++

			// Ensure Q2 capacity after promotion
			c.ensureAmCapacity()
			return entry.value, true
		}
	}

	// Check A1out (Qout) - counts as miss because value is not in memory
	if _, ok := c.outItems[key]; ok {
		c.misses++
		return nil, false
	}

	c.misses++
	return nil, false
}

// Set adds or updates an item in the cache.
// Following the 2Q algorithm:
// - If in Am or A1in: update value and move to MRU.
// - If in A1out: promote to Am.
// - If nowhere: add to A1in.
func (c *TwoQueueCache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. If key exists in memory (Q1 or Q2)
	if elem, ok := c.items[key]; ok {
		entry := elem.Value.(*twoQueueEntry) // safe: we control the type
		entry.value = value

		// Move to MRU of its respective list
		if entry.q == inQ2 {
			c.q2.MoveToFront(elem)
		} else {
			c.q1.MoveToFront(elem)
		}
		return
	}

	// 2. If key exists in A1out (Ghost history)
	if elem, ok := c.outItems[key]; ok {
		// Found in history, so we promote it directly to Am (Q2)
		c.qout.Remove(elem)
		delete(c.outItems, key)

		entry := &twoQueueEntry{key: key, value: value, q: inQ2}
		newElem := c.q2.PushFront(entry)
		c.items[key] = newElem
		c.promotionCount++

		c.ensureAmCapacity()
		return
	}

	// 3. New item: add to A1in (Q1)
	entry := &twoQueueEntry{key: key, value: value, q: inQ1}
	elem := c.q1.PushFront(entry)
	c.items[key] = elem

	c.ensureA1InCapacity()
}

// ensureA1InCapacity handles eviction from A1in to A1out.
func (c *TwoQueueCache) ensureA1InCapacity() {
	if c.q1.Len() > c.a1InCapacity {
		// Evict LRU from A1in
		elem := c.q1.Back()
		if elem != nil {
			c.q1.Remove(elem)
			entry := elem.Value.(*twoQueueEntry) // safe: we control the type
			delete(c.items, entry.key)

			// Transfer key to A1out history
			entry.q = inQout
			outElem := c.qout.PushFront(entry.key)
			c.outItems[entry.key] = outElem

			c.ensureA1OutCapacity()
		}
	}
}

// ensureAmCapacity handles eviction from Am.
func (c *TwoQueueCache) ensureAmCapacity() {
	if c.q2.Len() > c.amCapacity {
		// Evict LRU from Am (permanent deletion from memory)
		elem := c.q2.Back()
		if elem != nil {
			c.q2.Remove(elem)
			entry := elem.Value.(*twoQueueEntry) // safe: we control the type
			delete(c.items, entry.key)
		}
	}
}

// ensureA1OutCapacity handles eviction from Ghost history.
func (c *TwoQueueCache) ensureA1OutCapacity() {
	if c.qout.Len() > c.a1OutCapacity {
		// Evict oldest history key
		elem := c.qout.Back()
		if elem != nil {
			c.qout.Remove(elem)
			key := elem.Value.(string) // safe: we control the type
			delete(c.outItems, key)
		}
	}
}

// Delete removes an item from all queues.
func (c *TwoQueueCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		entry := elem.Value.(*twoQueueEntry) // safe: we control the type
		if entry.q == inQ1 {
			c.q1.Remove(elem)
		} else {
			c.q2.Remove(elem)
		}
		delete(c.items, key)
	}

	if elem, ok := c.outItems[key]; ok {
		c.qout.Remove(elem)
		delete(c.outItems, key)
	}
}

// Len returns the number of items currently held in memory.
func (c *TwoQueueCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Stats returns the performance metrics of the cache.
func (c *TwoQueueCache) Stats() (hits, misses, promotions int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses, c.promotionCount
}

// Purge completely clears the cache and metrics.
func (c *TwoQueueCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.q1.Init()
	c.q2.Init()
	c.qout.Init()

	c.items = make(map[string]*list.Element)
	c.outItems = make(map[string]*list.Element)

	c.hits = 0
	c.misses = 0
	c.promotionCount = 0
}
