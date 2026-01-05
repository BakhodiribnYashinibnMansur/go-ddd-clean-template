package cache

import (
	"gct/internal/repo/persistent/redis"
	pkgCache "gct/pkg/cache"
	"gct/pkg/logger"
)

type Cache struct {
	redis    *redis.Repo
	logger   logger.Log
	lru      *pkgCache.LRUCache
	memory   *pkgCache.MemoryCache
	fifo     *pkgCache.FIFOCache
	twoQueue *pkgCache.TwoQueueCache
	lifo     *pkgCache.LIFOCache
	random   *pkgCache.RandomCache
	mru      *pkgCache.MRUCache
	lfu      *pkgCache.LFUCache
	slru     *pkgCache.SLRUCache
	clock    *pkgCache.ClockCache
	jitter   *Jitter
}

func NewCache(redis *redis.Repo, logger logger.Log) *Cache {
	lruCache := pkgCache.NewLRUCache(1000)

	twoQueue := pkgCache.NewTwoQueueCache(1000)

	return &Cache{
		redis:    redis,
		logger:   logger,
		lru:      lruCache,
		memory:   pkgCache.NewMemoryCache(),
		fifo:     pkgCache.NewFIFOCache(1000),
		twoQueue: twoQueue,
		lifo:     pkgCache.NewLIFOCache(1000),
		random:   pkgCache.NewRandomCache(1000),
		mru:      pkgCache.NewMRUCache(1000),
		lfu:      pkgCache.NewLFUCache(1000),
		slru:     pkgCache.NewSLRUCache(1000, 0.8),
		clock:    pkgCache.NewClockCache(1000),
		jitter:   NewJitter(redis, logger),
	}
}
