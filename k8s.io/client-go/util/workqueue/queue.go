package workqueue

import (
	"sync"
	"time"

	"k8s.io/utils/clock"
)

type Interface interface {
	Add(item interface{})
	Len() int
	Get() (item interface{}, shutdown bool)
	Done(item interface{})
	ShutDown()
	ShutDownWithDrain()
	ShuttingDown() bool
}

func New() *Type {
	return NewNamed("")
}

func NewNamed(name string) *Type {
	rc := clock.RealClock{}

	return newQueue(
		rc,
		globalMetricsFactory.newQueueMetrics(name, rc),
		defaultUnfinishedWorkUpdatePeriod,
	)
}

func newQueue(c clock.WithTicker, metrics queueMetrics, updatePeriod time.Duration) *Type {
	t := &Type{
		clock:                      c,
		dirty:                      set{},
		processing:                 set{},
		cond:                       sync.NewCond(&sync.Mutex{}),
		metrics:                    metrics,
		unfinishedWorkUpdatePeriod: updatePeriod,
	}
	if _, ok := metrics.(noMetrics); !ok {
		go t.updateUnfinishedWorkLoop()
	}
	return t
}

const defaultUnfinishedWorkUpdatePeriod = 500 * time.Millisecond

// 实现Interface接口
type Type struct {
	queue []t
	// 需要被处理的条目
	dirty                      set
	processing                 set
	cond                       *sync.Cond
	shuttingDown               bool
	drain                      bool
	metrics                    queueMetrics
	unfinishedWorkUpdatePeriod time.Duration
	clock                      clock.WithTicker
}

type empty struct{}
type t interface{}
type set map[t]empty

func (s set) has(item t) bool {
	_, exists := s[item]
	return exists
}

func (s set) insert(item t) {
	s[item] = empty{}
}

func (s set) delete(item t) {
	delete(s, item)
}

func (s set) len() int {
	return len(s)
}

func (q *Type) Add(item interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	if q.shuttingDown {
		return
	}
	if q.dirty.has(item) {
		return
	}

	q.metrics.add(item)

	// 入队列的时候，将其标记为脏数据
	q.dirty.insert(item)
	// 如果正在处理，那么不会再入队
	if q.processing.has(item) {
		return
	}

	q.queue = append(q.queue, item)
	// 有新数据了，唤醒一个goroutine
	q.cond.Signal()
}

// not in queue -> in queue with dirty -> out dirty in processing -> out processing maybe in queue again when dirty
func (q *Type) Len() int {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return len(q.queue)
}

func (q *Type) Get() (item interface{}, shutdown bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	// 没有数据可读，也没有正在关闭，就等待被唤醒
	for len(q.queue) == 0 && !q.shuttingDown {
		q.cond.Wait()
	}
	// 唤醒后，发现还是没有数据可读，就直接返回
	if len(q.queue) == 0 {
		return nil, true
	}

	item = q.queue[0]
	q.queue[0] = nil
	q.queue = q.queue[1:]

	q.metrics.get(item)

	// 获取到数据后，标记为正在处理
	q.processing.insert(item)
	// 获取到了后，就从取消脏数据的标记
	q.dirty.delete(item)

	return item, false
}

func (q *Type) Done(item interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.metrics.done(item)

	q.processing.delete(item)
	// 处理完了之前还有相同的，就再次入队，等待被获取
	if q.dirty.has(item) {
		q.queue = append(q.queue, item)
		q.cond.Signal()
	} else if q.processing.len() == 0 {
		q.cond.Signal()
	}
}

func (q *Type) ShutDown() {
	q.setDrain(false)
	q.shutdown()
}

// 等待处理完了才会关闭
func (q *Type) ShutDownWithDrain() {
	q.setDrain(true)
	q.shutdown()
	for q.isProcessing() && q.shouldDrain() {
		q.waitForProcessing()
	}
}

func (q *Type) isProcessing() bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return q.processing.len() != 0
}

func (q *Type) waitForProcessing() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if q.processing.len() == 0 {
		return
	}
	q.cond.Wait()
}

func (q *Type) setDrain(shouldDrain bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	q.drain = shouldDrain
}

func (q *Type) shouldDrain() bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return q.drain
}

func (q *Type) shutdown() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	q.shuttingDown = true
	q.cond.Broadcast()
}

func (q *Type) ShuttingDown() bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	return q.shuttingDown
}

func (q *Type) updateUnfinishedWorkLoop() {
	t := q.clock.NewTicker(q.unfinishedWorkUpdatePeriod)
	defer t.Stop()
	for range t.C() {
		if !func() bool {
			q.cond.L.Lock()
			defer q.cond.L.Unlock()
			if !q.shuttingDown {
				q.metrics.updateUnfinishedWork()
				return true
			}
			return false
		}() {
			return
		}
	}
}
