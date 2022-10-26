package cache

import (
	"container/list"
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

type LRUExpireCache struct {
	// clock is used to obtain the current time
	clock Clock

	lock sync.Mutex

	maxSize      int
	evictionList list.List
	entries      map[interface{}]*list.Element
}

func NewLRUExpireCache(maxSize int) *LRUExpireCache {
	return NewLRUExpireCacheWithClock(maxSize, realClock{})
}

func NewLRUExpireCacheWithClock(maxSize int, clock Clock) *LRUExpireCache {
	if maxSize <= 0 {
		panic("maxSize must be > 0")
	}

	return &LRUExpireCache{
		clock:   clock,
		maxSize: maxSize,
		entries: map[interface{}]*list.Element{},
	}
}

type cacheEntry struct {
	key        interface{}
	value      interface{}
	expireTime time.Time
}

func (c *LRUExpireCache) Add(key interface{}, value interface{}, ttl time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()

	oldElement, ok := c.entries[key]
	if ok {
		c.evictionList.MoveToFront(oldElement)
		oldElement.Value.(*cacheEntry).value = value
		oldElement.Value.(*cacheEntry).expireTime = c.clock.Now().Add(ttl)
		return
	}

	if c.evictionList.Len() >= c.maxSize {
		toEvict := c.evictionList.Back()
		c.evictionList.Remove(toEvict)
		delete(c.entries, toEvict.Value.(*cacheEntry).key)
	}

	entry := &cacheEntry{
		key:        key,
		value:      value,
		expireTime: c.clock.Now().Add(ttl),
	}
	element := c.evictionList.PushFront(entry)
	c.entries[key] = element
}

func (c *LRUExpireCache) Get(key interface{}) (interface{}, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	element, ok := c.entries[key]
	if !ok {
		return nil, false
	}

	if c.clock.Now().After(element.Value.(*cacheEntry).expireTime) {
		c.evictionList.Remove(element)
		delete(c.entries, key)
		return nil, false
	}

	c.evictionList.MoveToFront(element)

	return element.Value.(*cacheEntry).value, true
}

func (c *LRUExpireCache) Remove(key interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	element, ok := c.entries[key]
	if !ok {
		return
	}

	c.evictionList.Remove(element)
	delete(c.entries, key)
}

func (c *LRUExpireCache) Keys() []interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()

	now := c.clock.Now()

	val := make([]interface{}, 0, c.evictionList.Len())
	for element := c.evictionList.Back(); element != nil; element = element.Prev() {
		// Only return unexpired keys
		if !now.After(element.Value.(*cacheEntry).expireTime) {
			val = append(val, element.Value.(*cacheEntry).key)
		}
	}

	return val
}
