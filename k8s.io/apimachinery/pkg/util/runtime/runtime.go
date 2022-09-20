package runtime

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

var (
	ReallyCrash = true
)

var PanicHandlers = []func(interface{}){logPanic}

func HandleCrash(additionalHandlers ...func(interface{})) {
	if r := recover(); r != nil {
		for _, fn := range PanicHandlers {
			fn(r)
		}
		for _, fn := range additionalHandlers {
			fn(r)
		}
		if ReallyCrash {
			panic(r)
		}
	}
}

func logPanic(r interface{}) {
	if r == http.ErrAbortHandler {
		return
	}

	const size = 64 << 10
	stacktrace := make([]byte, size)
	stacktrace = stacktrace[:runtime.Stack(stacktrace, false)]
	if _, ok := r.(string); ok {
		klog.Errorf("Observed a panic: %s\n%s", r, stacktrace)
	} else {
		klog.Errorf("Observed a panic: %#v (%v)\n%s", r, r, stacktrace)
	}
}

var ErrorHandlers = []func(error){
	logError,
	(&rudimentaryErrorBackoff{
		lastErrorTime: time.Now(),
		minPeriod:     time.Millisecond,
	}).OnError,
}

func HandleError(err error) {
	if err == nil {
		return
	}

	for _, fn := range ErrorHandlers {
		fn(err)
	}
}

func logError(err error) {
	klog.ErrorDepth(2, err)
}

type rudimentaryErrorBackoff struct {
	minPeriod         time.Duration // immutable
	lastErrorTimeLock sync.Mutex
	lastErrorTime     time.Time
}

func (r *rudimentaryErrorBackoff) OnError(error) {
	r.lastErrorTimeLock.Lock()
	defer r.lastErrorTimeLock.Unlock()
	d := time.Since(r.lastErrorTime)
	if d < r.minPeriod {
		time.Sleep(r.minPeriod - d)
	}
	r.lastErrorTime = time.Now()
}

// 返回调用该函数的函数名字
func GetCaller() string {
	var pc [1]uintptr
	runtime.Callers(3, pc[:])
	f := runtime.FuncForPC(pc[0])
	if f == nil {
		return fmt.Sprintf("Unable to find caller")
	}
	return f.Name()
}

func RecoverFromPanic(err *error) {
	if r := recover(); r != nil {
		const size = 64 << 10
		stacktrace := make([]byte, size)
		stacktrace = stacktrace[:runtime.Stack(stacktrace, false)]

		*err = fmt.Errorf(
			"recovered from panic %q. (err=%v) Call stack:\n%s",
			r,
			*err,
			stacktrace)
	}
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}
