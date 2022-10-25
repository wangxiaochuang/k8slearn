package cache

import (
	"hash/fnv"
	"time"
)

// split cache lookups across N striped caches
type stripedCache struct {
	stripeCount uint32
	hashFunc    func(string) uint32
	caches      []cache
}

type hashFunc func(string) uint32
type newCacheFunc func() cache

func newStripedCache(stripeCount int, hash hashFunc, newCacheFunc newCacheFunc) cache {
	caches := []cache{}
	for i := 0; i < stripeCount; i++ {
		caches = append(caches, newCacheFunc())
	}
	return &stripedCache{
		stripeCount: uint32(stripeCount),
		hashFunc:    hash,
		caches:      caches,
	}
}

func (c *stripedCache) get(key string) (*cacheRecord, bool) {
	return c.caches[c.hashFunc(key)%c.stripeCount].get(key)
}
func (c *stripedCache) set(key string, value *cacheRecord, ttl time.Duration) {
	c.caches[c.hashFunc(key)%c.stripeCount].set(key, value, ttl)
}
func (c *stripedCache) remove(key string) {
	c.caches[c.hashFunc(key)%c.stripeCount].remove(key)
}

func fnvHashFunc(key string) uint32 {
	f := fnv.New32()
	f.Write([]byte(key))
	return f.Sum32()
}
