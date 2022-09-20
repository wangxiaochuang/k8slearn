package rest

import (
	"sync"

	"k8s.io/klog/v2"
)

type WarningHandler interface {
	HandleWarningHeader(code int, agent string, text string)
}

var (
	defaultWarningHandler     WarningHandler = WarningLogger{}
	defaultWarningHandlerLock sync.RWMutex
)

func SetDefaultWarningHandler(l WarningHandler) {
	defaultWarningHandlerLock.Lock()
	defer defaultWarningHandlerLock.Unlock()
	defaultWarningHandler = l
}

type NoWarnings struct{}

func (NoWarnings) HandleWarningHeader(code int, agent string, message string) {}

type WarningLogger struct{}

func (WarningLogger) HandleWarningHeader(code int, agent string, message string) {
	if code != 299 || len(message) == 0 {
		return
	}
	klog.Warning(message)
}
