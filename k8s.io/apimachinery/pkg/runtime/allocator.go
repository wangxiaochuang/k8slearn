package runtime

import "sync"

var AllocatorPool = sync.Pool{
	New: func() interface{} {
		return &Allocator{}
	},
}

type Allocator struct {
	buf []byte
}

var _ MemoryAllocator = &Allocator{}

func (a *Allocator) Allocate(n uint64) []byte {
	if uint64(cap(a.buf)) >= n {
		a.buf = a.buf[:n]
		return a.buf
	}
	size := uint64(2*cap(a.buf)) + n
	a.buf = make([]byte, size, size)
	a.buf = a.buf[:n]
	return a.buf
}

type SimpleAllocator struct{}

var _ MemoryAllocator = &SimpleAllocator{}

func (sa *SimpleAllocator) Allocate(n uint64) []byte {
	return make([]byte, n, n)
}
