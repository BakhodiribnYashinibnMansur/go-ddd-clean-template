package cache

import (
	"gct/internal/repo/persistent/redis"
	"gct/pkg/logger"
	"go.uber.org/zap"
)

type Cache struct {
	redis    *redis.Repo
	logger   logger.Log
	lru      *LRUCache
	memory   *MemoryCache
	fifo     *FIFOCache
	twoQueue *TwoQueueCache
	lifo     *LIFOCache
	random   *RandomCache
	mru      *MRUCache
	lfu      *LFUCache
	slru     *SLRUCache
	clock    *ClockCache
}

func NewCache(redis *redis.Repo, logger logger.Log) *Cache {
	lruCache, err := NewLRUCache(1000)
	if err != nil {
		logger.Errorw("failed to create lru cache", zap.Error(err))
	}

	twoQueue, err := NewTwoQueueCache(1000)
	if err != nil {
		logger.Errorw("failed to create 2q cache", zap.Error(err))
	}

	return &Cache{
		redis:    redis,
		logger:   logger,
		lru:      lruCache,
		memory:   NewMemoryCache(),
		fifo:     NewFIFOCache(1000),
		twoQueue: twoQueue,
		lifo:     NewLIFOCache(1000),
		random:   NewRandomCache(1000),
		mru:      NewMRUCache(1000),
		lfu:      NewLFUCache(1000),
		slru:     NewSLRUCache(1000, 0.8),
		clock:    NewClockCache(1000),
	}
}
