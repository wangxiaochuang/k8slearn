package audit

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/errors"
	auditinternal "k8s.io/apiserver/pkg/apis/audit"
)

// Union returns an audit Backend which logs events to a set of backends. The returned
// Sink implementation blocks in turn for each call to ProcessEvents.
func Union(backends ...Backend) Backend {
	if len(backends) == 1 {
		return backends[0]
	}
	return union{backends}
}

type union struct {
	backends []Backend
}

func (u union) ProcessEvents(events ...*auditinternal.Event) bool {
	success := true
	for _, backend := range u.backends {
		success = backend.ProcessEvents(events...) && success
	}
	return success
}

func (u union) Run(stopCh <-chan struct{}) error {
	var funcs []func() error
	for _, backend := range u.backends {
		funcs = append(funcs, func() error {
			return backend.Run(stopCh)
		})
	}
	return errors.AggregateGoroutines(funcs...)
}

func (u union) Shutdown() {
	for _, backend := range u.backends {
		backend.Shutdown()
	}
}

func (u union) String() string {
	var backendStrings []string
	for _, backend := range u.backends {
		backendStrings = append(backendStrings, fmt.Sprintf("%s", backend))
	}
	return fmt.Sprintf("union[%s]", strings.Join(backendStrings, ","))
}
