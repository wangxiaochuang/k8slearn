package healthz

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

type HealthChecker interface {
	Name() string
	Check(req *http.Request) error
}

var PingHealthz HealthChecker = ping{}

type ping struct{}

func (ping) Name() string {
	return "ping"
}

func (ping) Check(_ *http.Request) error {
	return nil
}

var LogHealthz HealthChecker = &log{}

type log struct {
	startOnce    sync.Once
	lastVerified atomic.Value
}

func (l *log) Name() string {
	return "log"
}

func (l *log) Check(_ *http.Request) error {
	l.startOnce.Do(func() {
		l.lastVerified.Store(time.Now())
		go wait.Forever(func() {
			klog.Flush()
			l.lastVerified.Store(time.Now())
		}, time.Minute)
	})

	lastVerified := l.lastVerified.Load().(time.Time)
	if time.Since(lastVerified) < (2 * time.Minute) {
		return nil
	}
	return fmt.Errorf("logging blocked")
}
