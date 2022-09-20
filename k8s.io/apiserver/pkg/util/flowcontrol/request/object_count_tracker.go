package request

import (
	"errors"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
)

const (
	pruneInterval            = 1 * time.Hour
	staleTolerationThreshold = 3 * time.Minute
)

var (
	ObjectCountNotFoundErr = errors.New("object count not found for the given resource")

	ObjectCountStaleErr = errors.New("object count has gone stale for the given resource")
)

type StorageObjectCountTracker interface {
	Set(string, int64)
	Get(string) (int64, error)
}

func NewStorageObjectCountTracker(stopCh <-chan struct{}) StorageObjectCountTracker {
	tracker := &objectCountTracker{
		clock:  &clock.RealClock{},
		counts: map[string]*timestampedCount{},
	}
	go func() {
		// 每一个小时删除一次过期的内容
		wait.PollUntil(
			pruneInterval,
			func() (bool, error) {
				return false, tracker.prune(pruneInterval)
			}, stopCh)
		klog.InfoS("StorageObjectCountTracker pruner is exiting")
	}()

	return tracker
}

type timestampedCount struct {
	count         int64
	lastUpdatedAt time.Time
}

type objectCountTracker struct {
	clock clock.PassiveClock

	lock   sync.RWMutex
	counts map[string]*timestampedCount
}

func (t *objectCountTracker) Set(groupResource string, count int64) {
	if count <= -1 {
		return
	}

	now := t.clock.Now()

	t.lock.Lock()
	defer t.lock.Unlock()

	if item, ok := t.counts[groupResource]; ok {
		item.count = count
		item.lastUpdatedAt = now
		return
	}

	t.counts[groupResource] = &timestampedCount{
		count:         count,
		lastUpdatedAt: now,
	}
}

func (t *objectCountTracker) Get(groupResource string) (int64, error) {
	staleThreshold := t.clock.Now().Add(-staleTolerationThreshold)

	t.lock.RLock()
	defer t.lock.RUnlock()

	if item, ok := t.counts[groupResource]; ok {
		if item.lastUpdatedAt.Before(staleThreshold) {
			return item.count, ObjectCountStaleErr
		}
		return item.count, nil
	}
	return 0, ObjectCountNotFoundErr
}

// 过期了就删掉
func (t *objectCountTracker) prune(threshold time.Duration) error {
	oldestLastUpdatedAtAllowed := t.clock.Now().Add(-threshold)

	t.lock.Lock()
	defer t.lock.Unlock()

	for groupResource, count := range t.counts {
		if count.lastUpdatedAt.After(oldestLastUpdatedAtAllowed) {
			continue
		}
		delete(t.counts, groupResource)
	}

	return nil
}
