package cache

import (
	"gct/internal/repo/persistent/redis"
	pkgcache "gct/internal/shared/infrastructure/cache"
	"gct/internal/shared/infrastructure/logger"
)

type Cache struct {
	redis    *redis.Repo
	logger   logger.Log
	lru      *pkgcache.LRUCache
	memory   *pkgcache.MemoryCache
	fifo     *pkgcache.FIFOCache
	twoQueue *pkgcache.TwoQueueCache
	lifo     *pkgcache.LIFOCache
	random   *pkgcache.RandomCache
	mru      *pkgcache.MRUCache
	lfu      *pkgcache.LFUCache
	slru     *pkgcache.SLRUCache
	clock    *pkgcache.ClockCache
	jitter   *Jitter
}

func NewCache(redis *redis.Repo, logger logger.Log) *Cache {
	lruCache := pkgcache.NewLRUCache(1000)

	twoQueue := pkgcache.NewTwoQueueCache(1000)

	return &Cache{
		redis:    redis,
		logger:   logger,
		lru:      lruCache,
		memory:   pkgcache.NewMemoryCache(),
		fifo:     pkgcache.NewFIFOCache(1000),
		twoQueue: twoQueue,
		lifo:     pkgcache.NewLIFOCache(1000),
		random:   pkgcache.NewRandomCache(1000),
		mru:      pkgcache.NewMRUCache(1000),
		lfu:      pkgcache.NewLFUCache(1000),
		slru:     pkgcache.NewSLRUCache(1000, 0.8),
		clock:    pkgcache.NewClockCache(1000),
		jitter:   NewJitter(redis, logger),
	}
}
