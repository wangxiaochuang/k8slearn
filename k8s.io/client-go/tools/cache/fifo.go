package cache

import (
	"errors"
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"
)

type PopProcessFunc func(interface{}) error

type ErrRequeue struct {
	Err error
}

var ErrFIFOClosed = errors.New("DeltaFIFO: manipulating with closed queue")

func (e ErrRequeue) Error() string {
	if e.Err == nil {
		return "the popped item should be requeued without returning an error"
	}
	return e.Err.Error()
}

type Queue interface {
	Store
	Pop(PopProcessFunc) (interface{}, error)
	AddIfNotPresent(interface{}) error
	HasSynced() bool
	Close()
}

func Pop(queue Queue) interface{} {
	var result interface{}
	queue.Pop(func(obj interface{}) error {
		result = obj
		return nil
	})
	return result
}

type FIFO struct {
	lock                   sync.RWMutex
	cond                   sync.Cond
	items                  map[string]interface{}
	queue                  []string
	populated              bool
	initialPopulationCount int
	keyFunc                KeyFunc
	closed                 bool
}

var (
	_ = Queue(&FIFO{}) // FIFO is a Queue
)

func (f *FIFO) Close() {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.closed = true
	f.cond.Broadcast()
}

func (f *FIFO) HasSynced() bool {
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.populated && f.initialPopulationCount == 0
}

func (f *FIFO) Add(obj interface{}) error {
	//通过obj获取id
	id, err := f.keyFunc(obj)
	if err != nil {
		return KeyError{obj, err}
	}
	f.lock.Lock()
	defer f.lock.Unlock()
	f.populated = true
	if _, exists := f.items[id]; !exists {
		f.queue = append(f.queue, id)
	}
	f.items[id] = obj
	f.cond.Broadcast()
	return nil
}

func (f *FIFO) AddIfNotPresent(obj interface{}) error {
	id, err := f.keyFunc(obj)
	if err != nil {
		return KeyError{obj, err}
	}
	f.lock.Lock()
	defer f.lock.Unlock()
	f.addIfNotPresent(id, obj)
	return nil
}

func (f *FIFO) addIfNotPresent(id string, obj interface{}) {
	f.populated = true
	if _, exists := f.items[id]; exists {
		return
	}

	f.queue = append(f.queue, id)
	f.items[id] = obj
	f.cond.Broadcast()
}

func (f *FIFO) Update(obj interface{}) error {
	return f.Add(obj)
}

func (f *FIFO) Delete(obj interface{}) error {
	id, err := f.keyFunc(obj)
	if err != nil {
		return KeyError{obj, err}
	}
	f.lock.Lock()
	defer f.lock.Unlock()
	f.populated = true
	delete(f.items, id)
	return err
}

func (f *FIFO) List() []interface{} {
	f.lock.RLock()
	defer f.lock.RUnlock()
	list := make([]interface{}, 0, len(f.items))
	for _, item := range f.items {
		list = append(list, item)
	}
	return list
}

func (f *FIFO) ListKeys() []string {
	f.lock.RLock()
	defer f.lock.RUnlock()
	list := make([]string, 0, len(f.items))
	for key := range f.items {
		list = append(list, key)
	}
	return list
}

func (f *FIFO) Get(obj interface{}) (item interface{}, exists bool, err error) {
	key, err := f.keyFunc(obj)
	if err != nil {
		return nil, false, KeyError{obj, err}
	}
	return f.GetByKey(key)
}

func (f *FIFO) GetByKey(key string) (item interface{}, exists bool, err error) {
	f.lock.RLock()
	defer f.lock.RUnlock()
	item, exists = f.items[key]
	return item, exists, nil
}

func (f *FIFO) IsClosed() bool {
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.closed
}

func (f *FIFO) Pop(process PopProcessFunc) (interface{}, error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	for {
		for len(f.queue) > 0 {
			if f.closed {
				return nil, ErrFIFOClosed
			}
			f.cond.Wait() // 没有数据可读就等待
		}
		id := f.queue[0]
		f.queue = f.queue[1:]
		if f.initialPopulationCount > 0 {
			f.initialPopulationCount--
		}
		item, ok := f.items[id]
		if !ok {
			continue
		}
		delete(f.items, id)
		err := process(item)
		if e, ok := err.(ErrRequeue); ok {
			f.addIfNotPresent(id, item)
			err = e.Err
		}
		return item, err
	}
}

func (f *FIFO) Replace(list []interface{}, resourceVersion string) error {
	items := make(map[string]interface{}, len(list))
	for _, item := range list {
		key, err := f.keyFunc(item)
		if err != nil {
			return KeyError{item, err}
		}
		items[key] = item
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	if !f.populated {
		f.populated = true
		f.initialPopulationCount = len(items)
	}

	f.items = items
	f.queue = f.queue[:0]
	for id := range items {
		f.queue = append(f.queue, id)
	}
	if len(f.queue) > 0 {
		f.cond.Broadcast()
	}
	return nil
}

func (f *FIFO) Resync() error {
	f.lock.Lock()
	defer f.lock.Unlock()

	inQueue := sets.NewString()
	for _, id := range f.queue {
		inQueue.Insert(id)
	}
	for id := range f.items {
		if !inQueue.Has(id) {
			f.queue = append(f.queue, id)
		}
	}
	if len(f.queue) > 0 {
		f.cond.Broadcast()
	}
	return nil
}

func NewFIFO(keyFunc KeyFunc) *FIFO {
	f := &FIFO{
		items:   map[string]interface{}{},
		queue:   []string{},
		keyFunc: keyFunc,
	}
	f.cond.L = &f.lock
	return f
}
