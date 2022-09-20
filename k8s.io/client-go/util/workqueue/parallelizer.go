package workqueue

import (
	"context"
	"sync"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

type DoWorkPieceFunc func(piece int)

type options struct {
	chunkSize int
}

type Options func(*options)

func WithChunkSize(c int) func(*options) {
	return func(o *options) {
		o.chunkSize = c
	}
}

func ParallelizeUntil(ctx context.Context, workers, pieces int, doWorkPiece DoWorkPieceFunc, opts ...Options) {
	if pieces == 0 {
		return
	}
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}
	chunkSize := o.chunkSize
	if chunkSize < 1 {
		chunkSize = 1
	}

	chunks := ceilDiv(pieces, chunkSize)
	toProcess := make(chan int, chunks)
	for i := 0; i < chunks; i++ {
		toProcess <- i
	}
	close(toProcess)

	var stop <-chan struct{}
	if ctx != nil {
		stop = ctx.Done()
	}
	if chunks < workers {
		workers = chunks
	}
	wg := sync.WaitGroup{}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer utilruntime.HandleCrash()
			defer wg.Done()
			for chunk := range toProcess {
				start := chunk * chunkSize
				end := start + chunkSize
				if end > pieces {
					end = pieces
				}
				for p := start; p < end; p++ {
					select {
					case <-stop:
						return
					default:
						doWorkPiece(p)
					}
				}
			}
		}()
	}
	wg.Wait()
}

func ceilDiv(a, b int) int {
	return (a + b - 1) / b
}
