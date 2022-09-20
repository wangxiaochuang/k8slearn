package watch

import (
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

type Interface interface {
	Stop()
	ResultChan() <-chan Event
}

type EventType string

const (
	Added    EventType = "ADDED"
	Modified EventType = "MODIFIED"
	Deleted  EventType = "DELETED"
	Bookmark EventType = "BOOKMARK"
	Error    EventType = "ERROR"
)

var (
	DefaultChanSize int32 = 100
)

type Event struct {
	Type   EventType
	Object runtime.Object
}

type emptyWatch chan Event

func NewEmptyWatch() Interface {
	ch := make(chan Event)
	close(ch)
	return emptyWatch(ch)
}

func (w emptyWatch) Stop() {
}

func (w emptyWatch) ResultChan() <-chan Event {
	return chan Event(w)
}

type FakeWatcher struct {
	result  chan Event
	stopped bool
	sync.Mutex
}

func NewFake() *FakeWatcher {
	return &FakeWatcher{
		result: make(chan Event),
	}
}

func NewFakeWithChanSize(size int, blocking bool) *FakeWatcher {
	return &FakeWatcher{
		result: make(chan Event, size),
	}
}

func (f *FakeWatcher) Stop() {
	f.Lock()
	defer f.Unlock()
	if !f.stopped {
		klog.V(4).Infof("Stopping fake watcher.")
		close(f.result)
		f.stopped = true
	}
}

func (f *FakeWatcher) IsStopped() bool {
	f.Lock()
	defer f.Unlock()
	return f.stopped
}

func (f *FakeWatcher) Reset() {
	f.Lock()
	defer f.Unlock()
	f.stopped = false
	f.result = make(chan Event)
}

func (f *FakeWatcher) ResultChan() <-chan Event {
	return f.result
}

func (f *FakeWatcher) Add(obj runtime.Object) {
	f.result <- Event{Added, obj}
}

func (f *FakeWatcher) Modify(obj runtime.Object) {
	f.result <- Event{Modified, obj}
}

func (f *FakeWatcher) Delete(lastValue runtime.Object) {
	f.result <- Event{Deleted, lastValue}
}

func (f *FakeWatcher) Error(errValue runtime.Object) {
	f.result <- Event{Error, errValue}
}

func (f *FakeWatcher) Action(action EventType, obj runtime.Object) {
	f.result <- Event{action, obj}
}

type RaceFreeFakeWatcher struct {
	result  chan Event
	Stopped bool
	sync.Mutex
}

func NewRaceFreeFake() *RaceFreeFakeWatcher {
	return &RaceFreeFakeWatcher{
		result: make(chan Event, DefaultChanSize),
	}
}

func (f *RaceFreeFakeWatcher) Stop() {
	f.Lock()
	defer f.Unlock()
	if !f.Stopped {
		klog.V(4).Infof("Stopping fake watcher.")
		close(f.result)
		f.Stopped = true
	}
}

func (f *RaceFreeFakeWatcher) IsStopped() bool {
	f.Lock()
	defer f.Unlock()
	return f.Stopped
}

// Reset prepares the watcher to be reused.
func (f *RaceFreeFakeWatcher) Reset() {
	f.Lock()
	defer f.Unlock()
	f.Stopped = false
	f.result = make(chan Event, DefaultChanSize)
}

func (f *RaceFreeFakeWatcher) ResultChan() <-chan Event {
	f.Lock()
	defer f.Unlock()
	return f.result
}

func (f *RaceFreeFakeWatcher) Add(obj runtime.Object) {
	f.Lock()
	defer f.Unlock()
	if !f.Stopped {
		select {
		case f.result <- Event{Added, obj}:
			return
		default:
			panic(fmt.Errorf("channel full"))
		}
	}
}

func (f *RaceFreeFakeWatcher) Modify(obj runtime.Object) {
	f.Lock()
	defer f.Unlock()
	if !f.Stopped {
		select {
		case f.result <- Event{Modified, obj}:
			return
		default:
			panic(fmt.Errorf("channel full"))
		}
	}
}

func (f *RaceFreeFakeWatcher) Delete(lastValue runtime.Object) {
	f.Lock()
	defer f.Unlock()
	if !f.Stopped {
		select {
		case f.result <- Event{Deleted, lastValue}:
			return
		default:
			panic(fmt.Errorf("channel full"))
		}
	}
}

func (f *RaceFreeFakeWatcher) Error(errValue runtime.Object) {
	f.Lock()
	defer f.Unlock()
	if !f.Stopped {
		select {
		case f.result <- Event{Error, errValue}:
			return
		default:
			panic(fmt.Errorf("channel full"))
		}
	}
}

func (f *RaceFreeFakeWatcher) Action(action EventType, obj runtime.Object) {
	f.Lock()
	defer f.Unlock()
	if !f.Stopped {
		select {
		case f.result <- Event{action, obj}:
			return
		default:
			panic(fmt.Errorf("channel full"))
		}
	}
}

type ProxyWatcher struct {
	result chan Event
	stopCh chan struct{}

	mutex   sync.Mutex
	stopped bool
}

var _ Interface = &ProxyWatcher{}

func NewProxyWatcher(ch chan Event) *ProxyWatcher {
	return &ProxyWatcher{
		result:  ch,
		stopCh:  make(chan struct{}),
		stopped: false,
	}
}

func (pw *ProxyWatcher) Stop() {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	if !pw.stopped {
		pw.stopped = true
		close(pw.stopCh)
	}
}

func (pw *ProxyWatcher) Stopping() bool {
	pw.mutex.Lock()
	defer pw.mutex.Unlock()
	return pw.stopped
}

func (pw *ProxyWatcher) ResultChan() <-chan Event {
	return pw.result
}

func (pw *ProxyWatcher) StopChan() <-chan struct{} {
	return pw.stopCh
}
