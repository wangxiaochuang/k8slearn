package workqueue

import (
	"container/heap"
	"sync"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/clock"
)

type DelayingInterface interface {
	Interface

	AddAfter(item interface{}, duration time.Duration)
}

func NewDelayingQueue() DelayingInterface {
	return NewDelayingQueueWithCustomClock(clock.RealClock{}, "")
}

func NewDelayingQueueWithCustomQueue(q Interface, name string) DelayingInterface {
	return newDelayingQueue(clock.RealClock{}, q, name)
}

func NewNamedDelayingQueue(name string) DelayingInterface {
	return NewDelayingQueueWithCustomClock(clock.RealClock{}, name)
}

func NewDelayingQueueWithCustomClock(clock clock.WithTicker, name string) DelayingInterface {
	return newDelayingQueue(clock, NewNamed(name), name)
}

func newDelayingQueue(clock clock.WithTicker, q Interface, name string) *delayingType {
	ret := &delayingType{
		Interface:       q,
		clock:           clock,
		heartbeat:       clock.NewTicker(maxWait),
		stopCh:          make(chan struct{}),
		waitingForAddCh: make(chan *waitFor, 1000),
		metrics:         newRetryMetrics(name),
	}

	go ret.waitingLoop()
	return ret
}

type delayingType struct {
	Interface

	clock clock.Clock

	stopCh chan struct{}

	stopOnce sync.Once

	heartbeat clock.Ticker

	waitingForAddCh chan *waitFor

	metrics retryMetrics
}

type waitFor struct {
	data    t
	readyAt time.Time
	index   int
}

type waitForPriorityQueue []*waitFor

func (pq waitForPriorityQueue) Len() int {
	return len(pq)
}
func (pq waitForPriorityQueue) Less(i, j int) bool {
	return pq[i].readyAt.Before(pq[j].readyAt)
}
func (pq waitForPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *waitForPriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*waitFor)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *waitForPriorityQueue) Pop() interface{} {
	n := len(*pq)
	item := (*pq)[n-1]
	item.index = -1
	*pq = (*pq)[0:(n - 1)]
	return item
}

func (pq waitForPriorityQueue) Peek() interface{} {
	return pq[0]
}

func (q *delayingType) ShutDown() {
	q.stopOnce.Do(func() {
		q.Interface.ShutDown()
		close(q.stopCh)
		q.heartbeat.Stop()
	})
}

func (q *delayingType) AddAfter(item interface{}, duration time.Duration) {
	if q.ShuttingDown() {
		return
	}

	q.metrics.retry()

	if duration <= 0 {
		q.Add(item)
		return
	}

	// 没有结束的话，直接写到channel里，准备添加
	select {
	case <-q.stopCh:
	case q.waitingForAddCh <- &waitFor{data: item, readyAt: q.clock.Now().Add(duration)}:
	}
}

const maxWait = 10 * time.Second

func (q *delayingType) waitingLoop() {
	defer utilruntime.HandleCrash()

	never := make(<-chan time.Time)

	var nextReadyAtTimer clock.Timer

	waitingForQueue := &waitForPriorityQueue{}
	heap.Init(waitingForQueue)

	waitingEntryByData := map[t]*waitFor{}

	for {
		// 如果正在关闭，那延迟的所有待入队内容都不管了
		if q.Interface.ShuttingDown() {
			return
		}

		now := q.clock.Now()

		// 将到期的所有都入队
		for waitingForQueue.Len() > 0 {
			entry := waitingForQueue.Peek().(*waitFor)
			if entry.readyAt.After(now) {
				break
			}

			// 找到了一个到期的，直接放到队列里
			entry = heap.Pop(waitingForQueue).(*waitFor)
			q.Add(entry.data)
			delete(waitingEntryByData, entry.data)
		}

		nextReadyAt := never
		if waitingForQueue.Len() > 0 {
			if nextReadyAtTimer != nil {
				nextReadyAtTimer.Stop()
			}

			// 最近一个即将到期的条目
			entry := waitingForQueue.Peek().(*waitFor)
			nextReadyAtTimer = q.clock.NewTimer(entry.readyAt.Sub(now))
			nextReadyAt = nextReadyAtTimer.C()
		}

		select {
		case <-q.stopCh:
			return
			// 10s 有一次心跳
		case <-q.heartbeat.C():
		case <-nextReadyAt:
		case waitEntry := <-q.waitingForAddCh:
			if waitEntry.readyAt.After(q.clock.Now()) {
				// 还没到时间就添加到等待队列里，添加还会记录一下状态
				insert(waitingForQueue, waitingEntryByData, waitEntry)
			} else {
				q.Add(waitEntry.data)
			}

			drained := false
			for !drained {
				select {
				// 将channel里的都添加到等待队列
				case waitEntry := <-q.waitingForAddCh:
					if waitEntry.readyAt.After(q.clock.Now()) {
						insert(waitingForQueue, waitingEntryByData, waitEntry)
					} else {
						q.Add(waitEntry.data)
					}
				default:
					drained = true
				}
			}
		}
	}
}

func insert(q *waitForPriorityQueue, knownEntries map[t]*waitFor, entry *waitFor) {
	existing, exists := knownEntries[entry.data]
	// 已经添加过了还有添加一次，只修改时间
	if exists {
		if existing.readyAt.After(entry.readyAt) {
			existing.readyAt = entry.readyAt
			heap.Fix(q, existing.index)
		}

		return
	}

	heap.Push(q, entry)
	knownEntries[entry.data] = entry
}
