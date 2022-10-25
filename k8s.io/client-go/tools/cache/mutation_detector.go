package cache

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/diff"
	"k8s.io/klog/v2"
)

var mutationDetectionEnabled = false

func init() {
	mutationDetectionEnabled, _ = strconv.ParseBool(os.Getenv("KUBE_CACHE_MUTATION_DETECTOR"))
}

type MutationDetector interface {
	AddObject(obj interface{})
	Run(stopCh <-chan struct{})
}

func NewCacheMutationDetector(name string) MutationDetector {
	if !mutationDetectionEnabled {
		return dummyMutationDetector{}
	}
	klog.Warningln("Mutation detector is enabled, this will result in memory leakage.")
	return &defaultCacheMutationDetector{name: name, period: 1 * time.Second, retainDuration: 2 * time.Minute}
}

type dummyMutationDetector struct{}

func (dummyMutationDetector) Run(stopCh <-chan struct{}) {
}
func (dummyMutationDetector) AddObject(obj interface{}) {
}

type defaultCacheMutationDetector struct {
	name   string
	period time.Duration

	compareObjectsLock sync.Mutex

	addedObjsLock sync.Mutex
	addedObjs     []cacheObj

	cachedObjs []cacheObj

	retainDuration     time.Duration
	lastRotated        time.Time
	retainedCachedObjs []cacheObj

	failureFunc func(message string)
}

type cacheObj struct {
	cached interface{}
	copied interface{}
}

func (d *defaultCacheMutationDetector) Run(stopCh <-chan struct{}) {
	for {
		if d.lastRotated.IsZero() {
			d.lastRotated = time.Now()
		} else if time.Since(d.lastRotated) > d.retainDuration {
			d.retainedCachedObjs = d.cachedObjs
			d.cachedObjs = nil
			d.lastRotated = time.Now()
		}

		d.CompareObjects()

		select {
		case <-stopCh:
			return
		case <-time.After(d.period):
		}
	}
}

func (d *defaultCacheMutationDetector) AddObject(obj interface{}) {
	if _, ok := obj.(DeletedFinalStateUnknown); ok {
		return
	}
	if obj, ok := obj.(runtime.Object); ok {
		copiedObj := obj.DeepCopyObject()

		d.addedObjsLock.Lock()
		defer d.addedObjsLock.Unlock()
		d.addedObjs = append(d.addedObjs, cacheObj{cached: obj, copied: copiedObj})
	}
}

func (d *defaultCacheMutationDetector) CompareObjects() {
	d.compareObjectsLock.Lock()
	defer d.compareObjectsLock.Unlock()

	d.addedObjsLock.Lock()
	d.cachedObjs = append(d.cachedObjs, d.addedObjs...)
	d.addedObjs = nil
	d.addedObjsLock.Unlock()

	altered := false
	for i, obj := range d.cachedObjs {
		if !reflect.DeepEqual(obj.cached, obj.copied) {
			fmt.Printf("CACHE %s[%d] ALTERED!\n%v\n", d.name, i, diff.ObjectGoPrintSideBySide(obj.cached, obj.copied))
			altered = true
		}
	}
	for i, obj := range d.retainedCachedObjs {
		if !reflect.DeepEqual(obj.cached, obj.copied) {
			fmt.Printf("CACHE %s[%d] ALTERED!\n%v\n", d.name, i, diff.ObjectGoPrintSideBySide(obj.cached, obj.copied))
			altered = true
		}
	}
	if altered {
		msg := fmt.Sprintf("cache %s modified", d.name)
		if d.failureFunc != nil {
			d.failureFunc(msg)
			return
		}
		panic(msg)
	}
}
