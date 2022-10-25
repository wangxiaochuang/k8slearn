package cache

import (
	"time"

	utilcache "k8s.io/apimachinery/pkg/util/cache"
	"k8s.io/utils/clock"
)

type simpleCache struct {
	cache *utilcache.Expiring
}

func newSimpleCache(clock clock.Clock) cache {
	return &simpleCache{cache: utilcache.NewExpiringWithClock(clock)}
}

func (c *simpleCache) get(key string) (*cacheRecord, bool) {
	record, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}
	value, ok := record.(*cacheRecord)
	return value, ok
}

func (c *simpleCache) set(key string, value *cacheRecord, ttl time.Duration) {
	c.cache.Set(key, value, ttl)
}

func (c *simpleCache) remove(key string) {
	c.cache.Delete(key)
}
