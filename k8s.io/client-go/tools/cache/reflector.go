package cache

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"sync"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/naming"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
)

const defaultExpectedTypeName = "<unspecified>"

type Reflector struct {
	name             string
	expectedTypeName string
	expectedType     reflect.Type
	expectedGVK      *schema.GroupVersionKind
	store            Store
	listerWatcher    ListerWatcher

	backoffManager         wait.BackoffManager
	initConnBackoffManager wait.BackoffManager

	resyncPeriod                         time.Duration
	ShouldResync                         func() bool
	clock                                clock.Clock
	paginatedResult                      bool
	lastSyncResourceVersion              string
	isLastSyncResourceVersionUnavailable bool
	lastSyncResourceVersionMutex         sync.RWMutex
	WatchListPageSize                    int64
	watchErrorHandler                    WatchErrorHandler
}

// p123
type WatchErrorHandler func(r *Reflector, err error)

// p126
func DefaultWatchErrorHandler(r *Reflector, err error) {
	switch {
	case isExpiredError(err):
		// Don't set LastSyncResourceVersionUnavailable - LIST call with ResourceVersion=RV already
		// has a semantic that it returns data at least as fresh as provided RV.
		// So first try to LIST with setting RV to resource version of last observed object.
		klog.V(4).Infof("%s: watch of %v closed with: %v", r.name, r.expectedTypeName, err)
	case err == io.EOF:
		// watch closed normally
	case err == io.ErrUnexpectedEOF:
		klog.V(1).Infof("%s: Watch for %v closed with unexpected EOF: %v", r.name, r.expectedTypeName, err)
	default:
		utilruntime.HandleError(fmt.Errorf("%s: Failed to watch %v: %v", r.name, r.expectedTypeName, err))
	}
}

var (
	minWatchTimeout = 5 * time.Minute
)

// p166
func NewReflector(lw ListerWatcher, expectedType interface{}, store Store, resyncPeriod time.Duration) *Reflector {
	return NewNamedReflector(naming.GetNameFromCallsite(internalPackages...), lw, expectedType, store, resyncPeriod)
}

func NewNamedReflector(name string, lw ListerWatcher, expectedType interface{}, store Store, resyncPeriod time.Duration) *Reflector {
	realClock := &clock.RealClock{}
	r := &Reflector{
		name:                   name,
		listerWatcher:          lw,
		store:                  store,
		backoffManager:         wait.NewExponentialBackoffManager(800*time.Millisecond, 30*time.Second, 2*time.Minute, 2.0, 1.0, realClock),
		initConnBackoffManager: wait.NewExponentialBackoffManager(800*time.Millisecond, 30*time.Second, 2*time.Minute, 2.0, 1.0, realClock),
		resyncPeriod:           resyncPeriod,
		clock:                  realClock,
		watchErrorHandler:      WatchErrorHandler(DefaultWatchErrorHandler),
	}
	r.setExpectedType(expectedType)
	return r
}

func (r *Reflector) setExpectedType(expectedType interface{}) {
	r.expectedType = reflect.TypeOf(expectedType)
	if r.expectedType == nil {
		r.expectedTypeName = defaultExpectedTypeName
		return
	}

	r.expectedTypeName = r.expectedType.String()

	if obj, ok := expectedType.(*unstructured.Unstructured); ok {
		// Use gvk to check that watch event objects are of the desired type.
		gvk := obj.GroupVersionKind()
		if gvk.Empty() {
			klog.V(4).Infof("Reflector from %s configured with expectedType of *unstructured.Unstructured with empty GroupVersionKind.", r.name)
			return
		}
		r.expectedGVK = &gvk
		r.expectedTypeName = gvk.String()
	}
}

// p213
var internalPackages = []string{"client-go/tools/cache/"}

func (r *Reflector) Run(stopCh <-chan struct{}) {
	klog.V(3).Infof("Starting reflector %s (%s) from %s", r.expectedTypeName, r.resyncPeriod, r.name)
	wait.BackoffUntil(func() {
		if err := r.ListAndWatch(stopCh); err != nil {
			r.watchErrorHandler(r, err)
		}
	}, r.backoffManager, true, stopCh)
	klog.V(3).Infof("Stopping reflector %s (%s) from %s", r.expectedTypeName, r.resyncPeriod, r.name)
}

var (
	neverExitWatch     <-chan time.Time = make(chan time.Time)
	errorStopRequested                  = errors.New("stop requested")
)

func (r *Reflector) resyncChan() (<-chan time.Time, func() bool) {
	if r.resyncPeriod == 0 {
		return neverExitWatch, func() bool { return false }
	}
	t := r.clock.NewTimer(r.resyncPeriod)
	return t.C(), t.Stop
}

func (r *Reflector) ListAndWatch(stopCh <-chan struct{}) error {
	panic("not implemented")
}

// p542
func (r *Reflector) LastSyncResourceVersion() string {
	r.lastSyncResourceVersionMutex.RLock()
	defer r.lastSyncResourceVersionMutex.RUnlock()
	return r.lastSyncResourceVersion
}

// p585
func isExpiredError(err error) bool {
	return apierrors.IsResourceExpired(err) || apierrors.IsGone(err)
}
