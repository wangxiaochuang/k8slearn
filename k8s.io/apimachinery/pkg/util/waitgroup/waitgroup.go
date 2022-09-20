package waitgroup

import "sync"

type SafeWaitGroup struct {
	wg   sync.WaitGroup
	mu   sync.RWMutex
	wait bool
}
