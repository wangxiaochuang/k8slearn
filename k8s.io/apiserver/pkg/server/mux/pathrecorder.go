package mux

import (
	"net/http"
	"sync"
	"sync/atomic"
)

type PathRecorderMux struct {
	name            string
	lock            sync.Mutex
	notFoundHandler http.Handler
	pathToHandler   map[string]http.Handler
	prefixToHandler map[string]http.Handler
	mux             atomic.Value
	exposedPaths    []string
	pathStacks      map[string]string
}
