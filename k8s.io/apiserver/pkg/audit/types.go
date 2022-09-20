package audit

import (
	auditinternal "k8s.io/apiserver/pkg/apis/audit"
)

type Sink interface {
	ProcessEvents(events ...*auditinternal.Event) bool
}

type Backend interface {
	Sink
	Run(stopCh <-chan struct{}) error
	Shutdown()
	String() string
}
