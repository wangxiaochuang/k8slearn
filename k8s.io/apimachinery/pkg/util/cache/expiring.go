package cache

import (
	"container/heap"
	"sync"
	"time"

	"k8s.io/utils/clock"
)

// NewExpiring returns an initialized expiring cache.
func NewExpiring() *Expiring {
	return NewExpiringWithClock(clock.RealClock{})
}

// NewExpiringWithClock is like NewExpiring but allows passing in a custom
// clock for testing.
func NewExpiringWithClock(clock clock.Clock) *Expiring {
	return &Expiring{
		clock: clock,
		cache: make(map[interface{}]entry),
	}
}

type Expiring struct {
	clock clock.Clock

	// mu protects the below fields
	mu sync.RWMutex
	// cache is the internal map that backs the cache.
	cache      map[interface{}]entry
	generation uint64

	heap expiringHeap
}

type entry struct {
	val        interface{}
	expiry     time.Time
	generation uint64
}

func (c *Expiring) Get(key interface{}) (val interface{}, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.cache[key]
	if !ok || !c.clock.Now().Before(e.expiry) {
		return nil, false
	}
	return e.val, true
}

func (c *Expiring) Set(key interface{}, val interface{}, ttl time.Duration) {
	now := c.clock.Now()
	expiry := now.Add(ttl)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.generation++

	c.cache[key] = entry{
		val:        val,
		expiry:     expiry,
		generation: c.generation,
	}

	// Run GC inline before pushing the new entry.
	c.gc(now)

	heap.Push(&c.heap, &expiringHeapEntry{
		key:        key,
		expiry:     expiry,
		generation: c.generation,
	})
}

// Delete deletes an entry in the map.
func (c *Expiring) Delete(key interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.del(key, 0)
}

func (c *Expiring) del(key interface{}, generation uint64) {
	e, ok := c.cache[key]
	if !ok {
		return
	}
	if generation != 0 && generation != e.generation {
		return
	}
	delete(c.cache, key)
}

func (c *Expiring) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

func (c *Expiring) gc(now time.Time) {
	for {
		if len(c.heap) == 0 || now.Before(c.heap[0].expiry) {
			return
		}
		cleanup := heap.Pop(&c.heap).(*expiringHeapEntry)
		c.del(cleanup.key, cleanup.generation)
	}
}

type expiringHeapEntry struct {
	key        interface{}
	expiry     time.Time
	generation uint64
}

type expiringHeap []*expiringHeapEntry

var _ heap.Interface = &expiringHeap{}

func (cq expiringHeap) Len() int {
	return len(cq)
}

func (cq expiringHeap) Less(i, j int) bool {
	return cq[i].expiry.Before(cq[j].expiry)
}

func (cq expiringHeap) Swap(i, j int) {
	cq[i], cq[j] = cq[j], cq[i]
}

func (cq *expiringHeap) Push(c interface{}) {
	*cq = append(*cq, c.(*expiringHeapEntry))
}

func (cq *expiringHeap) Pop() interface{} {
	c := (*cq)[cq.Len()-1]
	*cq = (*cq)[:cq.Len()-1]
	return c
}
