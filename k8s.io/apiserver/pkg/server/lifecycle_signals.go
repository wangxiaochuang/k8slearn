package server

import "sync"

// p91
type lifecycleSignal interface {
	Signal()

	Signaled() <-chan struct{}

	Name() string
}

// p112
type lifecycleSignals struct {
	ShutdownInitiated          lifecycleSignal
	AfterShutdownDelayDuration lifecycleSignal
	InFlightRequestsDrained    lifecycleSignal
	HTTPServerStoppedListening lifecycleSignal
	HasBeenReady               lifecycleSignal
	MuxAndDiscoveryComplete    lifecycleSignal
}

// p142
func newLifecycleSignals() lifecycleSignals {
	return lifecycleSignals{
		ShutdownInitiated:          newNamedChannelWrapper("ShutdownInitiated"),
		AfterShutdownDelayDuration: newNamedChannelWrapper("AfterShutdownDelayDuration"),
		InFlightRequestsDrained:    newNamedChannelWrapper("InFlightRequestsDrained"),
		HTTPServerStoppedListening: newNamedChannelWrapper("HTTPServerStoppedListening"),
		HasBeenReady:               newNamedChannelWrapper("HasBeenReady"),
		MuxAndDiscoveryComplete:    newNamedChannelWrapper("MuxAndDiscoveryComplete"),
	}
}

func newNamedChannelWrapper(name string) lifecycleSignal {
	return &namedChannelWrapper{
		name: name,
		once: sync.Once{},
		ch:   make(chan struct{}),
	}
}

type namedChannelWrapper struct {
	name string
	once sync.Once
	ch   chan struct{}
}

func (e *namedChannelWrapper) Signal() {
	e.once.Do(func() {
		close(e.ch)
	})
}

func (e *namedChannelWrapper) Signaled() <-chan struct{} {
	return e.ch
}

func (e *namedChannelWrapper) Name() string {
	return e.name
}
