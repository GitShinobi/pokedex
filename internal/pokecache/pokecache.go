package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	entry map[string]cacheEntry
	mu    *sync.Mutex
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(interval time.Duration) *Cache {
	entry := make(map[string]cacheEntry)
	mu := sync.Mutex{}
	c := Cache{entry: entry, mu: &mu}
	go c.reapLoop(interval)
	return &c
}

func (c *Cache) Add(key string, val []byte) {
	(*c).mu.Lock()
	c.entry[key] = cacheEntry{time.Now(), val}
	(*c).mu.Unlock()
}

func (c *Cache) Get(key string) ([]byte, bool) {
	(*c).mu.Lock()
	entry, ok := (*c).entry[key]
	(*c).mu.Unlock()

	return entry.val, ok
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		(*c).mu.Lock()
		for k, cache := range (*c).entry {
			if time.Since(cache.createdAt) >= interval {
				delete((*c).entry, k)
			}
		}
		(*c).mu.Unlock()

	}
}
