package cache

import (
	"container/heap"
	"sync"
	"time"
)

// LFUCache implements Least Frequently Used cache
type LFUCache struct {
	capacity int
	items    map[string]*lfuItem
	pq       priorityQueue
	mu       sync.RWMutex
}

type lfuItem struct {
	key       string
	value     any
	frequency int
	index     int   // index in heap
	timestamp int64 // for tie-breaking (LRU within LFU)
}

// NewLFUCache creates a new LFU cache
func NewLFUCache(capacity int) *LFUCache {
	return &LFUCache{
		capacity: capacity,
		items:    make(map[string]*lfuItem),
		pq:       make(priorityQueue, 0, capacity),
	}
}

// Set adds a value
func (c *LFUCache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, ok := c.items[key]; ok {
		item.value = value
		item.frequency++
		item.timestamp = time.Now().UnixNano()
		heap.Fix(&c.pq, item.index)
		return
	}

	if len(c.items) >= c.capacity {
		c.evict()
	}

	item := &lfuItem{
		key:       key,
		value:     value,
		frequency: 1,
		timestamp: time.Now().UnixNano(),
	}
	c.items[key] = item
	heap.Push(&c.pq, item)
}

// Get retrieves a value
func (c *LFUCache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, ok := c.items[key]; ok {
		item.frequency++
		item.timestamp = time.Now().UnixNano()
		heap.Fix(&c.pq, item.index)
		return item.value, true
	}
	return nil, false
}

// Remove removes a value
func (c *LFUCache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, ok := c.items[key]; ok {
		heap.Remove(&c.pq, item.index)
		delete(c.items, key)
	}
}

func (c *LFUCache) evict() {
	if c.pq.Len() == 0 {
		return
	}
	item := heap.Pop(&c.pq).(*lfuItem)
	delete(c.items, item.key)
}

// Len returns size
func (c *LFUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Purge clears cache
func (c *LFUCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*lfuItem)
	c.pq = make(priorityQueue, 0, c.capacity)
}

// PriorityQueue implementation
type priorityQueue []*lfuItem

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	// If frequencies equal, use timestamp (LRU tie-breaker)
	if pq[i].frequency == pq[j].frequency {
		return pq[i].timestamp < pq[j].timestamp
	}
	return pq[i].frequency < pq[j].frequency
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*lfuItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}
