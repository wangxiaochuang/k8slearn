package server

import restclient "k8s.io/client-go/rest"

type PostStartHookFunc func(context PostStartHookContext) error

type PreShutdownHookFunc func() error

type PostStartHookContext struct {
	LoopbackClientConfig *restclient.Config
	StopCh               <-chan struct{}
}

type PostStartHookConfigEntry struct {
	hook             PostStartHookFunc
	originatingStack string
}
